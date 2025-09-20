// k6 run -e CONSUL_HTTP_ADDR=http://127.0.0.1:8500 -e CONSUL_SERVICE=emshop-coupon-srv -e FLASH_SALE_ID=101 -e USER_ID=20001 scripts/k6/grpc_coupon.js

import http from 'k6/http';
import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

const client = new grpc.Client();
client.load(['../../api/coupon/v1'], 'coupon.proto');

const CONSUL_HTTP_ADDR = __ENV.CONSUL_HTTP_ADDR || 'http://localhost:8500';
const CONSUL_SERVICE = __ENV.CONSUL_SERVICE || 'emshop-coupon-srv';
const CONSUL_TAG = __ENV.CONSUL_TAG || '';
const FLASH_SALE_ID = Number(__ENV.FLASH_SALE_ID) || 1;
const USER_ID = Number(__ENV.USER_ID) || 10001;
const ORDER_AMOUNT = Number(__ENV.ORDER_AMOUNT) || 199.0;
const COUPON_IDS = (__ENV.COUPON_IDS || '1').split(',').map((id) => Number(id.trim()) || 1);

export const options = {
  scenarios: {
    flashsale_stock: {
      executor: 'constant-arrival-rate',
      exec: 'flashSaleStock',
      rate: Number(__ENV.FLASH_STOCK_RPS) || 800,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 100,
      maxVUs: Number(__ENV.MAX_VUS) || 1500,
    },
    flashsale_participate: {
      executor: 'constant-arrival-rate',
      exec: 'flashSaleParticipate',
      rate: Number(__ENV.FLASH_PARTICIPATE_RPS) || 1200,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 150,
      maxVUs: Number(__ENV.MAX_VUS) || 3000,
      startTime: '10s',
    },
    coupon_calculate: {
      executor: 'constant-arrival-rate',
      exec: 'calculateDiscount',
      rate: Number(__ENV.COUPON_CALC_RPS) || 400,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 80,
      maxVUs: Number(__ENV.MAX_VUS) || 1500,
      startTime: '20s',
    },
  },
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
  const host = entry.Service.Address || entry.Node.Address;
  const port = entry.Service.Port;
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

function withClient(target, fn) {
  client.connect(target, { plaintext: true });
  try {
    fn();
  } finally {
    client.close();
  }
}

export function flashSaleStock(data) {
  withClient(data.target, () => {
    const res = client.invoke('proto.Coupon/GetFlashSaleStock', {
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
    const res = client.invoke('proto.Coupon/ParticipateFlashSale', {
      userId: USER_ID,
      flashSaleId: FLASH_SALE_ID,
    });
    check(res, {
      'participation accepted': (r) => r && r.status === grpc.StatusOK,
    });
  });
  sleep(__ENV.SLEEP || 0.1);
}

export function calculateDiscount(data) {
  withClient(data.target, () => {
    const res = client.invoke('proto.Coupon/CalculateCouponDiscount', {
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
