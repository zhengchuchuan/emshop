/*
  - 直连验证（先绕开 Consul，确认逻辑正确）
    k6 run -e GRPC_COUPON_TARGET=127.0.0.1:28056 -e USER_ID_MODE=PER_ITER -e FLASH_PARTICIPATE_RPS=800 -e PRE_VUS=40 -e MAX_VUS=400 scripts/k6/grpc_coupon.js
  - Consul 验证（启用后）
      - 确保 coupon 服务已读取更新后的 configs/coupon/srv.yaml 并重启注册
      k6 run \
            -e CONSUL_HTTP_ADDR=http://127.0.0.1:8500 \
            -e CONSUL_SERVICE=emshop-coupon-srv \
            -e USER_ID_MODE=PER_ITER \
            -e FLASH_PARTICIPATE_RPS=800 \
            -e PRE_VUS=40 \
            -e MAX_VUS=400 \
            scripts/k6/grpc_coupon.js


  k6 run \
  -e CONSUL_HTTP_ADDR=http://127.0.0.1:8500 \
  -e CONSUL_SERVICE=emshop-coupon-srv \
  -e FLASH_SALE_ID=1 \
  -e USER_ID_MODE=PER_VU \
  -e USER_ID_BASE=1000000 \
  -e FLASH_PARTICIPATE_RPS=5000 \
  -e PRE_VUS=500 \
  -e MAX_VUS=5000 \
  -e DURATION=30s \
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
    'checks{scenario:flashsale_participate}': ['rate>0.98'],
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
    client.connect(target, { plaintext: true });
    withClient._connected = true;
    withClient._target = target;
  }
  fn();
}

export function flashSaleStock(data) {
  withClient(data.target, () => {
    const res = client.invoke('Coupon/GetFlashSaleStock', {
      flashSaleId: FLASH_SALE_ID,
    });
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
    });
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
    });
    check(res, {
      'calc ok': (r) => r && r.status === grpc.StatusOK,
    });
  });
  sleep(__ENV.SLEEP || 0.1);
}
