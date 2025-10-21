/*
脚本用途：压测 Coupon gRPC 接口（参与秒杀/库存查询/优惠计算）。

快速执行（推荐，仅压参与秒杀）
  k6 run \
    -e GRPC_COUPON_TARGET=127.0.0.1:28056 \
    -e FLASH_SALE_ID=1 \
    -e USER_ID_MODE=PER_ITER \
    -e FLASH_PARTICIPATE_RPS=3000 \
    -e PRE_VUS=100 \
    -e MAX_VUS=3000 \
    -e GRPC_TIMEOUT=3s \
    -e ENABLE_FLASH_STOCK=0 \
    -e ENABLE_COUPON_CALC=0 \
    -e DURATION=120s \
    -e SUMMARY_HTML=./k6-summary.html \
    scripts/k6/grpc_coupon.js

更多用法（可按需调整 RPS）：
1) 仅压参与秒杀（直连，超时 3s，PER_ITER 生成唯一用户）：
   k6 run \
     -e GRPC_COUPON_TARGET=127.0.0.1:28056 \
     -e FLASH_SALE_ID=1 \
     -e USER_ID_MODE=PER_ITER \
     -e FLASH_PARTICIPATE_RPS=800 \
     -e PRE_VUS=60 \
     -e MAX_VUS=800 \
     -e GRPC_TIMEOUT=3s \
     -e ENABLE_FLASH_STOCK=0 \
     -e ENABLE_COUPON_CALC=0 \
     -e DURATION=120s \
     scripts/k6/grpc_coupon.js

2) 同时压三种场景（可按需调整 RPS）：
   k6 run \
     -e GRPC_COUPON_TARGET=127.0.0.1:28056 \
     -e FLASH_SALE_ID=1 \
     -e FLASH_STOCK_RPS=800 \
     -e FLASH_PARTICIPATE_RPS=5000 \
     -e COUPON_CALC_RPS=400 \
     -e PRE_VUS=80 \
     -e MAX_VUS=2000 \
     -e GRPC_TIMEOUT=3s \
     -e DURATION=120s \
     scripts/k6/grpc_coupon.js

3) 使用 Consul 解析（不直连时）：
   k6 run \
     -e CONSUL_HTTP_ADDR=http://127.0.0.1:8500 \
     -e CONSUL_SERVICE=emshop-coupon-srv \
     -e FLASH_SALE_ID=1 \
     -e FLASH_PARTICIPATE_RPS=800 \
     -e PRE_VUS=60 \
     -e MAX_VUS=800 \
     -e DURATION=120s \
     scripts/k6/grpc_coupon.js
*/
import http from 'k6/http';
import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const client = new grpc.Client();
client.load(['../../api/coupon/v1'], 'coupon.proto');

const CONSUL_HTTP_ADDR = __ENV.CONSUL_HTTP_ADDR || 'http://localhost:8500';
const CONSUL_SERVICE = __ENV.CONSUL_SERVICE || 'emshop-coupon-srv';
const CONSUL_TAG = __ENV.CONSUL_TAG || '';
const FLASH_SALE_ID = Number(__ENV.FLASH_SALE_ID) || 1;
// 用户ID生成策略：
// 1) 提供 USER_IDS（逗号分隔）时，从给定ID列表轮询；
// 2) 否则按 USER_ID_MODE 分配（PER_VU: 基于 __VU；PER_ITER: 基于 __ITER；CONST: 固定 USER_ID 或 USER_ID_BASE）
const USER_ID = Number(__ENV.USER_ID) || 0; // 当 CONST 且未给 USER_IDS 时使用
const USER_ID_BASE = Number(__ENV.USER_ID_BASE) || 100000;
const USER_ID_MODE = (__ENV.USER_ID_MODE || 'PER_VU').toUpperCase(); // PER_VU|PER_ITER|CONST
const USER_IDS = (__ENV.USER_IDS || '')
  .split(',')
  .map((s) => Number(s.trim()))
  .filter((n) => Number.isFinite(n) && n > 0);
const ORDER_AMOUNT = Number(__ENV.ORDER_AMOUNT) || 199.0;
const COUPON_IDS = (__ENV.COUPON_IDS || '1').split(',').map((id) => Number(id.trim()) || 1);

function buildScenarios() {
  const scenarios = {};
  const enableStock = (__ENV.ENABLE_FLASH_STOCK || '1') === '1';
  const enableParticipate = (__ENV.ENABLE_FLASH_PARTICIPATE || '1') === '1';
  const enableCalc = (__ENV.ENABLE_COUPON_CALC || '1') === '1';
  if (enableStock) {
    scenarios.flashsale_stock = {
      executor: 'constant-arrival-rate',
      exec: 'flashSaleStock',
      rate: Number(__ENV.FLASH_STOCK_RPS) || 800,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 20,
      maxVUs: Number(__ENV.MAX_VUS) || 200,
    };
  }
  if (enableParticipate) {
    scenarios.flashsale_participate = {
      executor: 'constant-arrival-rate',
      exec: 'flashSaleParticipate',
      rate: Number(__ENV.FLASH_PARTICIPATE_RPS) || 1200,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 40,
      maxVUs: Number(__ENV.MAX_VUS) || 400,
      startTime: '10s',
    };
  }
  if (enableCalc) {
    scenarios.coupon_calculate = {
      executor: 'constant-arrival-rate',
      exec: 'calculateDiscount',
      rate: Number(__ENV.COUPON_CALC_RPS) || 400,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 20,
      maxVUs: Number(__ENV.MAX_VUS) || 200,
      startTime: '20s',
    };
  }
  return scenarios;
}

export const options = {
  scenarios: buildScenarios(),
  thresholds: {
    // 关键场景的可用性与延迟门限（可按需调整）
    'checks{scenario:flashsale_participate}': ['rate>0.98'],
    'grpc_req_duration': ['p(95)<800'], // 95 分位 < 800ms
  },
};

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
  const explicitTarget = __ENV.GRPC_COUPON_TARGET;
  const target = explicitTarget && explicitTarget.length > 0
    ? explicitTarget
    : resolveFromConsul(CONSUL_SERVICE, CONSUL_TAG);
  console.log(`k6 gRPC target resolved to ${target}`);
  return { target };
}

// 自定义指标：仅在 gRPC OK 且业务成功(status=1)时计成功，否则计失败
const flashSuccess = new Counter('flash_sale_success_total');
const flashFail = new Counter('flash_sale_fail_total');

function pickUserId() {
  if (USER_IDS.length > 0) {
    // 轮询给定列表
    const idx = (Number(__VU) + Number(__ITER)) % USER_IDS.length;
    return USER_IDS[idx];
  }
  switch (USER_ID_MODE) {
    case 'CONST':
      return USER_ID > 0 ? USER_ID : USER_ID_BASE;
    case 'PER_ITER':
      return USER_ID_BASE + Number(__VU) * 1000000 + Number(__ITER); // 避免碰撞
    case 'PER_VU':
    default:
      return USER_ID_BASE + Number(__VU);
  }
}

function withClient(target, fn) {
  // 长连接：避免每次调用反复握手，降低端口耗尽风险
  if (!withClient._connected || withClient._target !== target) {
    try { client.close(); } catch (e) { /* ignore */ }
    client.connect(target, { plaintext: true, timeout: '5s' });
    withClient._connected = true;
    withClient._target = target;
  }
  fn();
}

export function flashSaleStock(data) {
  withClient(data.target, () => {
    const res = client.invoke('Coupon/GetFlashSaleStock', {
      flashSaleId: FLASH_SALE_ID,
    }, { timeout: __ENV.GRPC_TIMEOUT || '2s' });
    check(res, {
      'stock fetched': (r) => r && r.status === grpc.StatusOK,
      'has remaining': (r) => r && r.message && typeof r.message.remainingStock === 'number',
    });
  });
  sleep(__ENV.SLEEP || 0.1);
}

export function flashSaleParticipate(data) {
  withClient(data.target, () => {
    const userId = pickUserId();
    const res = client.invoke('Coupon/ParticipateFlashSale', {
      userId: userId,
      flashSaleId: FLASH_SALE_ID,
    }, { timeout: __ENV.GRPC_TIMEOUT || '3s' });
    const ok = check(res, {
      'grpc ok': (r) => r && r.status === grpc.StatusOK,
    });
    if (ok) {
      const msg = res.message || {};
      if (msg.status === 1) {
        flashSuccess.add(1);
      } else {
        flashFail.add(1);
      }
    } else {
      flashFail.add(1);
    }
  });
  sleep(__ENV.SLEEP || 0.1);
}

export function calculateDiscount(data) {
  withClient(data.target, () => {
    const res = client.invoke('Coupon/CalculateCouponDiscount', {
      userId: USER_ID,
      couponIds: COUPON_IDS,
      orderAmount: ORDER_AMOUNT,
      orderItems: [
        { goodsId: Number(__ENV.GOODS_ID) || 1, quantity: 1, price: ORDER_AMOUNT },
      ],
    }, { timeout: __ENV.GRPC_TIMEOUT || '2s' });
    check(res, {
      'calc ok': (r) => r && r.status === grpc.StatusOK,
    });
  });
  sleep(__ENV.SLEEP || 0.1);
}

// 自定义总结：输出 HTML 和 JSON（可通过环境变量关闭/改变路径）
export function handleSummary(data) {
  const out = {};
  const jsonPath = __ENV.SUMMARY_JSON || '';
  if (jsonPath) {
    out[jsonPath] = JSON.stringify(data);
  }
  // 控制台简要输出关键配置与阈值
  out['stdout'] = JSON.stringify({
    scenarios: Object.keys(options.scenarios || {}),
    rps: {
      stock: Number(__ENV.FLASH_STOCK_RPS) || 0,
      participate: Number(__ENV.FLASH_PARTICIPATE_RPS) || 0,
      calc: Number(__ENV.COUPON_CALC_RPS) || 0,
    },
    thresholds: options.thresholds || {},
  }, null, 2);
  return out;
}
