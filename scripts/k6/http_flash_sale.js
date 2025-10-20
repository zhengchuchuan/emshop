/*
用法示例：
1) 直接指定 BASE_URL 与 TOKEN（推荐在本地或已知网关地址）
   k6 run \
     -e BASE_URL=http://127.0.0.1:8051 \
     -e TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOjEsImF1dGhvcml0eV9pZCI6MiwiaXNzIjoiZW1zaG9wLWFwaSIsImV4cCI6MTc2NzA3NzU4NywibmJmIjoxNzU4NDM3NTg3LCJpYXQiOjE3NTg0Mzc1ODd9.gFzQAlukFxprFZyTxrZikACw7lTjaYuTekKqVoi-Dz4 \
     -e FLASH_SALE_ID=1 \
     -e RPS=2000 -e DURATION=30s -e PRE_VUS=200 -e MAX_VUS=2000 \
     scripts/k6/http_flash_sale.js

2) 通过 Consul 发现 emshop-api 实例（无需手写 BASE_URL）
   k6 run \
     -e CONSUL_HTTP_ADDR=http://127.0.0.1:8500 \
     -e CONSUL_SERVICE=emshop-api \
     -e TOKEN=your_jwt_token \
     -e RPS=2000 -e DURATION=30s \
     scripts/k6/http_flash_sale.js

补充：如果你的服务没有使用 "/api" 前缀，请设置 API_PREFIX=""（默认会自动探测 /api 与空前缀）。
*/

import http from 'k6/http';
import { check, group } from 'k6';
import { Counter } from 'k6/metrics';

// 自定义指标：记录业务成功/失败次数（HTTP 200 下：status=1 视为成功，=2 视为业务失败）
const flash_success = new Counter('flash_sale_success_total');
const flash_fail = new Counter('flash_sale_fail_total');

const CONSUL_HTTP_ADDR = __ENV.CONSUL_HTTP_ADDR || '';
const CONSUL_SERVICE = __ENV.CONSUL_SERVICE || 'emshop-api';
const CONSUL_TAG = __ENV.CONSUL_TAG || '';
const SERVICE_SCHEME = __ENV.SERVICE_SCHEME || 'http';
const API_PREFIX = __ENV.API_PREFIX || ''; // 若设置为"/api"或""则强制使用；默认自动探测

export const options = {
  scenarios: {
    flashsale_participate: {
      executor: 'constant-arrival-rate',
      exec: 'participate',
      rate: Number(__ENV.RPS) || 2000, // 目标到达率（即目标 QPS）
      timeUnit: '1s',
      duration: __ENV.DURATION || '30s',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 200,
      maxVUs: Number(__ENV.MAX_VUS) || 5000,
    },
  },
  thresholds: {
    'http_req_failed{scenario:flashsale_participate}': ['rate<0.01'],
    'http_req_duration{scenario:flashsale_participate}': ['p(95)<300', 'p(99)<600'],
    'checks{scenario:flashsale_participate}': ['rate>0.95'],
  },
  discardResponseBodies: true,
};

function resolveFromConsul(service, tag) {
  if (!CONSUL_HTTP_ADDR) {
    throw new Error('CONSUL_HTTP_ADDR 未设置且未提供 BASE_URL');
  }
  const url = `${CONSUL_HTTP_ADDR}/v1/health/service/${service}?passing=true`;
  const res = http.get(url);
  if (res.status !== 200) {
    throw new Error(`Consul 查询失败: ${res.status} ${res.body}`);
  }
  const instances = res.json();
  if (!instances || instances.length === 0) {
    throw new Error(`Consul 未发现可用实例: ${service}`);
  }
  const entry = tag
    ? instances.find((it) => it.Service && it.Service.Tags && it.Service.Tags.includes(tag))
    : instances[0];
  if (!entry || !entry.Service) {
    throw new Error(`Consul 返回数据异常: ${service}`);
  }
  const host = entry.Service.Address || entry.Node.Address;
  const port = entry.Service.Port;
  if (!host || !port) {
    throw new Error(`Consul 解析到无效地址: ${host}:${port}`);
  }
  return `${host}:${port}`;
}

function joinBase(origin, prefix) {
  const o = origin.replace(/\/$/, '');
  if (!prefix) return o;
  const p = prefix.replace(/^\/+|\/+$/g, '');
  return `${o}/${p}`;
}

function ensureAuthHeaders() {
  const token = __ENV.TOKEN || '';
  if (!token) {
    throw new Error('缺少 TOKEN（JWT）环境变量，无法调用受鉴权保护的接口');
  }
  return {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`,
  };
}

function tryActive(baseUrl) {
  const url = `${baseUrl}/v1/coupons/flash-sale/active`;
  const res = http.get(url, { tags: { name: 'flashsale_active' } });
  return { res, url };
}

function detectApiBase(origin) {
  if (API_PREFIX !== '') {
    return joinBase(origin, API_PREFIX);
  }
  // 自动探测：先试带 /api，再试空前缀
  const a = joinBase(origin, '/api');
  let { res } = tryActive(a);
  if (res && res.status === 200) return a;

  const b = joinBase(origin, '');
  ({ res } = tryActive(b));
  if (res && res.status === 200) return b;

  throw new Error('无法探测 API 前缀，请确认服务已启动，或通过 API_PREFIX=/api 显式指定');
}

function pickFlashSaleId(baseUrl) {
  // 若显式指定则使用之
  if (__ENV.FLASH_SALE_ID) {
    const id = Number(__ENV.FLASH_SALE_ID);
    if (!Number.isFinite(id) || id <= 0) throw new Error('FLASH_SALE_ID 非法');
    return id;
  }
  // 否则自动拉取进行中的活动
  const url = `${baseUrl}/v1/coupons/flash-sale/active`;
  const res = http.get(url, { tags: { name: 'flashsale_active' } });
  if (res.status !== 200) {
    throw new Error(`获取进行中秒杀失败: ${res.status} ${res.body}`);
  }
  const body = res.json();
  const items = (body && (body.items || body.data || [])) || [];
  if (!Array.isArray(items) || items.length === 0) {
    throw new Error('暂无进行中的秒杀活动，请通过 FLASH_SALE_ID 指定活动ID');
  }
  const id = items[0] && (items[0].id || items[0].ID || items[0].flash_sale_id);
  if (!id) throw new Error('秒杀活动列表解析失败，未获取到ID');
  return Number(id);
}

export function setup() {
  const explicitBase = __ENV.BASE_URL && __ENV.BASE_URL.length > 0 ? __ENV.BASE_URL : '';
  const origin = explicitBase || `${SERVICE_SCHEME}://${resolveFromConsul(CONSUL_SERVICE, CONSUL_TAG)}`;
  const baseUrl = detectApiBase(origin);
  console.log(`HTTP target resolved to ${baseUrl} (origin=${origin}, apiPrefix=${API_PREFIX || 'auto'})`);

  const headers = ensureAuthHeaders();
  const flashSaleId = pickFlashSaleId(baseUrl);
  console.log(`Using flash sale id = ${flashSaleId}`);

  return { baseUrl, headers, flashSaleId };
}

export function participate(data) {
  group('POST /v1/coupons/flash-sale/:id/participate', () => {
    const url = `${data.baseUrl}/v1/coupons/flash-sale/${data.flashSaleId}/participate`;
    const res = http.post(url, null, {
      headers: data.headers,
      tags: { name: 'flashsale_participate' },
    });

    const ok = check(res, {
      'http 200': (r) => r.status === 200,
    });

    // 仅在 HTTP 200 下统计业务成功/失败
    if (ok) {
      try {
        const body = res.json();
        const status = body && body.status;
        if (status === 1) {
          flash_success.add(1);
        } else {
          flash_fail.add(1);
        }
      } catch (e) {
        // 忽略 JSON 解析错误，按失败计入
        flash_fail.add(1);
      }
    }
  });
}
