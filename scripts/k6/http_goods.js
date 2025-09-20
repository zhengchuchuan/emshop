// k6 run -e CONSUL_HTTP_ADDR=http://127.0.0.1:8500 -e CONSUL_SERVICE=emshop-api scripts/k6/http_goods.js

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const LIST_QUERY = __ENV.GOODS_LIST_QUERY || 'pages=1&pagePerNums=50';
const CONSUL_HTTP_ADDR = __ENV.CONSUL_HTTP_ADDR || 'http://localhost:8500';
const CONSUL_SERVICE = __ENV.CONSUL_SERVICE || 'emshop-api';
const CONSUL_TAG = __ENV.CONSUL_TAG || '';
const SERVICE_SCHEME = __ENV.SERVICE_SCHEME || 'http';

const DEFAULT_SLEEP = Number(__ENV.SLEEP || 0.1);

export const options = {
  scenarios: {
    goods_list: {
      executor: 'constant-arrival-rate',
      exec: 'hitGoodsList',
      rate: Number(__ENV.GOODS_LIST_RPS) || 800,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 100,
      maxVUs: Number(__ENV.MAX_VUS) || 2000,
    },
    goods_detail: {
      executor: 'constant-arrival-rate',
      exec: 'hitGoodsDetail',
      rate: Number(__ENV.GOODS_DETAIL_RPS) || 1500,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 100,
      maxVUs: Number(__ENV.MAX_VUS) || 2000,
      startTime: '10s',
    },
    goods_stock: {
      executor: 'constant-arrival-rate',
      exec: 'hitGoodsStock',
      rate: Number(__ENV.GOODS_STOCK_RPS) || 1000,
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 100,
      maxVUs: Number(__ENV.MAX_VUS) || 2000,
      startTime: '20s',
    },
  },
  thresholds: {
    'http_req_duration{scenario:goods_list}': ['p(95)<300', 'p(99)<500'],
    'http_req_duration{scenario:goods_detail}': ['p(95)<350', 'p(99)<600'],
    'http_req_duration{scenario:goods_stock}': ['p(95)<250', 'p(99)<450'],
    'checks{scenario:goods_detail}': ['rate>0.99'],
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

function ensureApiBase(url) {
  if (url.endsWith('/api')) {
    return url;
  }
  return `${url.replace(/\/$/, '')}/api`;
}

function bootstrapGoods(baseUrl) {
  const listUrl = `${baseUrl}/v1/goods?${LIST_QUERY}`;
  const res = http.get(listUrl, { tags: { name: 'bootstrap_goods_list' } });
  if (!res || res.status !== 200) {
    console.error(`Failed to bootstrap goods list from ${listUrl}, status ${res && res.status}`);
    return [1];
  }
  const payload = res.json();
  const data = (payload && (payload.data || payload.goods || payload.items)) || [];
  const ids = data
    .map((item) => (item && (item.id || item.ID || item.goodsId)))
    .filter((id) => id !== undefined && id !== null);
  if (ids.length === 0) {
    console.warn('Bootstrap goods list is empty, falling back to ID=1.');
    return [1];
  }
  return ids;
}

export function setup() {
  const explicitBase = __ENV.BASE_URL && __ENV.BASE_URL.length > 0 ? __ENV.BASE_URL : '';
  const baseNoApi = explicitBase
    ? explicitBase
    : `${SERVICE_SCHEME}://${resolveFromConsul(CONSUL_SERVICE, CONSUL_TAG)}`;
  const baseUrl = ensureApiBase(baseNoApi);
  console.log(`HTTP target resolved to ${baseUrl}`);
  const goodsIds = bootstrapGoods(baseUrl);
  return { baseUrl, goodsIds };
}

function pickGoodsId(goodsIds) {
  if (!goodsIds || goodsIds.length === 0) {
    return 1;
  }
  return randomItem(goodsIds);
}

export function hitGoodsList(data) {
  group('GET /v1/goods', () => {
    const res = http.get(`${data.baseUrl}/v1/goods?${LIST_QUERY}`, {
      tags: { name: 'goods_list' },
    });
    check(res, {
      'status 200': (r) => r.status === 200,
      'has data array': (r) => {
        const body = r.json();
        const goods = body && (body.data || body.goods || body.items);
        return Array.isArray(goods);
      },
    });
  });
  sleep(DEFAULT_SLEEP);
}

export function hitGoodsDetail(data) {
  const goodsId = pickGoodsId(data.goodsIds);
  group('GET /v1/goods/:id', () => {
    const res = http.get(`${data.baseUrl}/v1/goods/${goodsId}`, {
      tags: { name: 'goods_detail' },
    });
    check(res, {
      'status 200': (r) => r.status === 200,
      'has id': (r) => {
        const body = r.json();
        return body && (body.id !== undefined || body.ID !== undefined);
      },
    });
  });
  sleep(DEFAULT_SLEEP);
}

export function hitGoodsStock(data) {
  const goodsId = pickGoodsId(data.goodsIds);
  group('GET /v1/goods/:id/stocks', () => {
    const res = http.get(`${data.baseUrl}/v1/goods/${goodsId}/stocks`, {
      tags: { name: 'goods_stock' },
    });
    check(res, {
      'status 200': (r) => r.status === 200,
      'has stock': (r) => {
        const body = r.json();
        return body && body.stocks !== undefined;
      },
    });
  });
  sleep(DEFAULT_SLEEP);
}
