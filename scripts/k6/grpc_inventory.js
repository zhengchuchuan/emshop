// k6 run -e CONSUL_HTTP_ADDR=http://127.0.0.1:8500 -e CONSUL_SERVICE=emshop-inventory-srv scripts/k6/grpc_inventory.js

import http from 'k6/http';
import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

const client = new grpc.Client();
client.load(['../../api/inventory/v1'], 'inventory.proto');

const CONSUL_HTTP_ADDR = __ENV.CONSUL_HTTP_ADDR || 'http://localhost:8500';
const CONSUL_SERVICE = __ENV.CONSUL_SERVICE || 'emshop-inventory-srv';
const CONSUL_TAG = __ENV.CONSUL_TAG || '';
const GOODS_IDS = (__ENV.GOODS_IDS || '1,2,3').split(',').map((id) => Number(id.trim()) || 1);

export const options = {
  scenarios: {
    inv_detail: {
      executor: 'constant-arrival-rate',
      exec: 'invDetail',
      rate: Number(__ENV.INV_DETAIL_RPS) || 600,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 50,
      maxVUs: Number(__ENV.MAX_VUS) || 500,
    },
    inv_sell: {
      executor: 'constant-arrival-rate',
      exec: 'invSell',
      rate: Number(__ENV.INV_SELL_RPS) || 300,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 50,
      maxVUs: Number(__ENV.MAX_VUS) || 500,
      startTime: '10s',
    },
    inv_reback: {
      executor: 'constant-arrival-rate',
      exec: 'invReback',
      rate: Number(__ENV.INV_REBACK_RPS) || 300,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 50,
      maxVUs: Number(__ENV.MAX_VUS) || 500,
      startTime: '20s',
    },
  },
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
  const host = entry.Service.Address || entry.Node.Address;
  const port = entry.Service.Port;
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
  client.connect(target, { plaintext: true });
  try {
    fn();
  } finally {
    client.close();
  }
}

export function invDetail(data) {
  const goodsId = pickGoodsId();
  withClient(data.target, () => {
    const res = client.invoke('proto.Inventory/InvDetail', { goodsId, num: 0 });
    check(res, {
      'status OK': (r) => r && r.status === grpc.StatusOK,
      'has inventory': (r) => r && r.message && typeof r.message.num === 'number',
    });
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
    const res = client.invoke('proto.Inventory/Sell', body);
    check(res, {
      'sell accepted': (r) => r && r.status === grpc.StatusOK,
    });
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
    const res = client.invoke('proto.Inventory/Reback', body);
    check(res, {
      'reback accepted': (r) => r && r.status === grpc.StatusOK,
    });
  });
  sleep(__ENV.SLEEP || 0.1);
}
