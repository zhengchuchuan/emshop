/*
用法示例（优先直连 gRPC，便于排查连通/超时）：

1) 直连 gRPC（连接复用 + 超时配置）：
   k6 run \
     -e GRPC_INVENTORY_TARGET=127.0.0.1:28057 \
     -e GOODS_IDS=1,2,3 \
     -e INV_DETAIL_RPS=600 \
     -e INV_SELL_RPS=300 \
     -e INV_REBACK_RPS=300 \
     -e PRE_VUS=60 \
     -e MAX_VUS=600 \
     -e DURATION=120s \
     -e GRPC_TIMEOUT=3s \
     -e GRPC_CONNECT_TIMEOUT=5s \
     scripts/k6/grpc_inventory.js

2) Consul 解析（不直连时）：
   k6 run \
     -e CONSUL_HTTP_ADDR=http://127.0.0.1:8500 \
     -e CONSUL_SERVICE=emshop-inventory-srv \
     -e GOODS_IDS=1,2,3 \
     -e PRE_VUS=60 \
     -e MAX_VUS=600 \
     -e DURATION=120s \
     scripts/k6/grpc_inventory.js

可选：快速禁用某些场景
  -e ENABLE_INV_DETAIL=0 或 ENABLE_INV_SELL=0 或 ENABLE_INV_REBACK=0
*/

import http from 'k6/http';
import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { Counter } from 'k6/metrics';

const client = new grpc.Client();
client.load(['../../api/inventory/v1'], 'inventory.proto');

const CONSUL_HTTP_ADDR = __ENV.CONSUL_HTTP_ADDR || 'http://localhost:8500';
const CONSUL_SERVICE = __ENV.CONSUL_SERVICE || 'emshop-inventory-srv';
const CONSUL_TAG = __ENV.CONSUL_TAG || '';
const GOODS_IDS = (__ENV.GOODS_IDS || '1,2,3').split(',').map((id) => Number(id.trim()) || 1);
const GRPC_TIMEOUT = __ENV.GRPC_TIMEOUT || '3s';
const GRPC_CONNECT_TIMEOUT = __ENV.GRPC_CONNECT_TIMEOUT || '5s';

// 场景启用开关（默认全开）
const ENABLE_INV_DETAIL = (__ENV.ENABLE_INV_DETAIL || '1') === '1';
const ENABLE_INV_SELL = (__ENV.ENABLE_INV_SELL || '1') === '1';
const ENABLE_INV_REBACK = (__ENV.ENABLE_INV_REBACK || '1') === '1';

function buildScenarios() {
  const scenarios = {};
  if (ENABLE_INV_DETAIL) {
    scenarios.inv_detail = {
      executor: 'constant-arrival-rate',
      exec: 'invDetail',
      rate: Number(__ENV.INV_DETAIL_RPS) || 600,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 50,
      maxVUs: Number(__ENV.MAX_VUS) || 500,
    };
  }
  if (ENABLE_INV_SELL) {
    scenarios.inv_sell = {
      executor: 'constant-arrival-rate',
      exec: 'invSell',
      rate: Number(__ENV.INV_SELL_RPS) || 300,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 50,
      maxVUs: Number(__ENV.MAX_VUS) || 500,
      startTime: '10s',
    };
  }
  if (ENABLE_INV_REBACK) {
    scenarios.inv_reback = {
      executor: 'constant-arrival-rate',
      exec: 'invReback',
      rate: Number(__ENV.INV_REBACK_RPS) || 300,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 50,
      maxVUs: Number(__ENV.MAX_VUS) || 500,
      startTime: '20s',
    };
  }
  return scenarios;
}

export const options = {
  scenarios: buildScenarios(),
  thresholds: {
    'checks{scenario:inv_detail}': ['rate>0.99'],
  },
};

const orderSeq = new SharedArray('order-counter', () => [{ n: 0 }]);

function nextOrderSn() {
  const counter = orderSeq[0];
  counter.n += 1;
  return `VT-${Date.now()}-${counter.n}`;
}

function pickGoodsId() {
  return GOODS_IDS[Math.floor(Math.random() * GOODS_IDS.length)] || 1;
}

function resolveFromConsul(service, tag) {
  const url = `${CONSUL_HTTP_ADDR}/v1/health/service/${service}?passing=true`;
  const res = http.get(url);
  if (res.status !== 200) {
    throw new Error(`Consul lookup failed: ${res.status} ${res.body}`);
  }
  const instances = res.json();
  if (!instances || instances.length === 0) {
    throw new Error(`No healthy instances found for service ${service}`);
  }
  const entry = tag
    ? instances.find((item) => item.Service && item.Service.Tags && item.Service.Tags.includes(tag))
    : instances[0];
  if (!entry || !entry.Service) {
    throw new Error(`Consul entry missing Service data for ${service}`);
  }
  const svc = entry.Service;
  if (svc.TaggedAddresses && svc.TaggedAddresses.grpc && svc.TaggedAddresses.grpc.Address) {
    const ep = String(svc.TaggedAddresses.grpc.Address);
    const m = ep.match(/^[a-z]+:\/\/([^/?#]+)(?:\?.*)?$/i);
    if (m && m[1]) {
      return m[1];
    }
  }
  const host = svc.Address || (entry.Node && entry.Node.Address);
  const port = svc.Port;
  if (!host || !port) {
    throw new Error(`Invalid address/port resolved for ${service}: ${host}:${port}`);
  }
  return `${host}:${port}`;
}

export function setup() {
  const explicitTarget = __ENV.GRPC_INVENTORY_TARGET;
  const target = explicitTarget && explicitTarget.length > 0
    ? explicitTarget
    : resolveFromConsul(CONSUL_SERVICE, CONSUL_TAG);
  console.log(`k6 gRPC target resolved to ${target}`);
  return { target };
}

function withClient(target, fn) {
  if (!withClient._connected || withClient._target !== target) {
    try { client.close(); } catch (e) { /* ignore */ }
    client.connect(target, { plaintext: true, timeout: GRPC_CONNECT_TIMEOUT });
    withClient._connected = true;
    withClient._target = target;
  }
  fn();
}

// 指标：按 gRPC 成功/失败计数
const invOk = new Counter('inventory_ok_total');
const invFail = new Counter('inventory_fail_total');

export function invDetail(data) {
  const goodsId = pickGoodsId();
  withClient(data.target, () => {
    try {
      const res = client.invoke('Inventory/InvDetail', { goodsId, num: 0 }, { timeout: GRPC_TIMEOUT });
      const ok = check(res, {
        'status OK': (r) => r && r.status === grpc.StatusOK,
      });
      if (ok) invOk.add(1); else invFail.add(1);
    } catch (e) {
      invFail.add(1);
      // 常见报错提示：method not found/连接错误
      console.error(`invDetail error: ${String(e)}`);
    }
  });
  sleep(__ENV.SLEEP || 0.1);
}

export function invSell(data) {
  const goodsId = pickGoodsId();
  const body = {
    goodsInfo: [{ goodsId, num: Number(__ENV.SELL_NUM) || 1 }],
    orderSn: nextOrderSn(),
  };
  withClient(data.target, () => {
    try {
      const res = client.invoke('Inventory/Sell', body, { timeout: GRPC_TIMEOUT });
      const ok = check(res, { 'sell accepted': (r) => r && r.status === grpc.StatusOK });
      if (ok) invOk.add(1); else invFail.add(1);
    } catch (e) {
      invFail.add(1);
      console.error(`invSell error: ${String(e)}`);
    }
  });
  sleep(__ENV.SLEEP || 0.1);
}

export function invReback(data) {
  const goodsId = pickGoodsId();
  const body = {
    goodsInfo: [{ goodsId, num: Number(__ENV.REBACK_NUM) || 1 }],
    orderSn: nextOrderSn(),
  };
  withClient(data.target, () => {
    try {
      const res = client.invoke('Inventory/Reback', body, { timeout: GRPC_TIMEOUT });
      const ok = check(res, { 'reback accepted': (r) => r && r.status === grpc.StatusOK });
      if (ok) invOk.add(1); else invFail.add(1);
    } catch (e) {
      invFail.add(1);
      console.error(`invReback error: ${String(e)}`);
    }
  });
  sleep(__ENV.SLEEP || 0.1);
}
