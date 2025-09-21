import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  vus: 20,
  duration: '30s',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

const BASE = __ENV.EMSHOP_BASE || 'http://127.0.0.1:8051';

export default function () {
  // 热点接口：券模板列表
  http.get(`${BASE}/v1/coupons/templates?page=1&pageSize=10`);
  // 其他接口可按需解开
  // http.get(`${BASE}/v1/goods`);
  // http.get(`${BASE}/v1/brands`);

  sleep(0.1);
}

