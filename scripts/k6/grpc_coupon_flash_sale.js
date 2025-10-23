/*
精简版：只压测 Coupon 秒杀 gRPC 接口（ParticipateFlashSale）。

推荐执行命令（直接指定速率，CAR 模式）：
  k6 run \
    -e GRPC_COUPON_TARGET=127.0.0.1:28056 \
    -e FLASH_SALE_ID=1 \
    -e FLASH_PARTICIPATE_RPS=4000 \
    -e PRE_VUS=400 -e MAX_VUS=4000 \
    -e GRPC_TIMEOUT=10s -e DURATION=15s \
    scripts/k6/grpc_coupon_flash_sale.js

按 VUS×PER_VU_RPS 计算速率（同样走 CAR）：
  k6 run \
    -e GRPC_COUPON_TARGET=127.0.0.1:28056 \
    -e FLASH_SALE_ID=1 \
    -e VUS=400 -e PER_VU_RPS=10 -e CVUS_MODE=1 \
    -e GRPC_TIMEOUT=10s -e DURATION=15s \
    scripts/k6/grpc_coupon_flash_sale.js

脚本封装（等价）：
  scripts/k6/run_flashsale_rate.sh
*/
// 引入 k6 的 gRPC 客户端模块（用于发起 gRPC 调用）
import grpc from 'k6/net/grpc';
// 引入 k6 的断言与节流工具：
// - check: 用于对返回结果做布尔判断，便于统计“通过/失败”的比例
// - sleep: 让 VU 暂停一段时间（秒），控制节奏
import { check, sleep } from 'k6';
// 引入自定义指标 Counter：自增计数器，用来记录成功/失败次数等
import { Counter } from 'k6/metrics';
// 友好汇总输出与 HTML 报告（显示完整统计）
// 引入人类友好的汇总渲染工具：用于最终输出在控制台（stdout）或生成 HTML
// 不使用外部 jslib，避免离线/受限网络失败

// 创建一个 gRPC 客户端对象（每个 VU 会共享此对象，但连接在 withClient 中维护）
const client = new grpc.Client();
// 加载 coupon 服务的 proto 定义：
// 第一个参数是 proto 搜索路径数组，第二个参数是 proto 文件名
client.load(['../../api/coupon/v1'], 'coupon.proto');

// 从环境变量读取活动 ID（字符串转数字）；未提供时默认 1
const FLASH_SALE_ID = Number(__ENV.FLASH_SALE_ID) || 1;
// 用户ID生成策略：
// 1) 提供 USER_IDS（逗号分隔）时，从给定ID列表轮询；
// 2) 否则按 USER_ID_MODE 分配（PER_VU: 基于 __VU；PER_ITER: 基于 __ITER；CONST: 固定 USER_ID 或 USER_ID_BASE）
// 如果选择 CONST 模式且未给 USER_IDS，则使用 USER_ID；未给则回退到 USER_ID_BASE
const USER_ID = Number(__ENV.USER_ID) || 0; // 当 CONST 且未给 USER_IDS 时使用
// 生成用户 ID 的基础值（避免与真实用户冲突）
const USER_ID_BASE = Number(__ENV.USER_ID_BASE) || 100000;
// 用户 ID 生成模式：PER_VU | PER_ITER | CONST（不区分大小写，这里统一大写）
const USER_ID_MODE = (__ENV.USER_ID_MODE || 'PER_VU').toUpperCase(); // PER_VU|PER_ITER|CONST
// 逗号分隔的用户 ID 列表（优先使用），过滤掉非法条目
const USER_IDS = (__ENV.USER_IDS || '')
  .split(',')
  .map((s) => Number(s.trim()))
  .filter((n) => Number.isFinite(n) && n > 0);
// 动态场景：
// - 当提供 FLASH_PARTICIPATE_RPS (>0) 时，使用 constant-arrival-rate（精准 RPS）
// - 否则当 CVUS_MODE=1 时，用 VUS×PER_VU_RPS 计算 rate 并同样使用 constant-arrival-rate（直指定速率）
// - 其它情况回退到 constant-vus
function buildOptions() {
  const vus = Number(__ENV.VUS) || 200;
  const perVu = Number(__ENV.PER_VU_RPS) || 10;
  const duration = __ENV.DURATION || '30s';
  const carRate = Number(__ENV.FLASH_PARTICIPATE_RPS) || 0;
  const useCVUSRate = String(__ENV.CVUS_MODE || '0') === '1';
  const rate = carRate > 0 ? carRate : (useCVUSRate ? vus * perVu : 0);
  if (rate > 0) {
    // 以固定到达速率驱动，每次迭代只发 1 次请求
    const preVUs = Number(__ENV.PRE_VUS) || vus;
    const maxVUs = Number(__ENV.MAX_VUS) || Math.max(preVUs, vus);
    return {
      scenarios: {
        flashsale_rate: {
          executor: 'constant-arrival-rate',
          rate: rate,
          timeUnit: '1s',
          duration: duration,
          preAllocatedVUs: preVUs,
          maxVUs: maxVUs,
          exec: 'flashSaleOnce',
        },
      },
      summaryTrendStats: ['avg', 'min', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
      discardResponseBodies: true,
    };
  }
  // 回退：constant-vus（按每 VU 目标速率节流）
  return {
    scenarios: {
      flashsale_cvus: {
        executor: 'constant-vus',
        exec: 'flashSaleLoop',
        vus: vus,
        duration: duration,
      },
    },
    summaryTrendStats: ['avg', 'min', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
    discardResponseBodies: true,
  };
}

export const options = buildOptions();

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
  const target = __ENV.GRPC_COUPON_TARGET || '127.0.0.1:28056'; // 读取目标地址
  console.log(`k6 gRPC target resolved to ${target}`);          // 打印确认
  return { target };                                            // 作为 data 传入执行函数
}

// 自定义指标：仅在 gRPC OK 且业务成功(status=1)时计成功，否则计失败
const flashSuccess = new Counter('flash_sale_success_total');   // 统计业务成功（status==1）
const flashFail = new Counter('flash_sale_fail_total');         // 统计失败（网络/超时/业务失败）

function pickUserId() {
  if (USER_IDS.length > 0) {
    // 若提供了用户列表，则按“VU编号+迭代次数”的和取模轮询
    const idx = (Number(__VU) + Number(__ITER)) % USER_IDS.length;
    return USER_IDS[idx];
  }
  switch (USER_ID_MODE) {
    case 'CONST':
      return USER_ID > 0 ? USER_ID : USER_ID_BASE;              // 固定用户
    case 'PER_ITER':
      return USER_ID_BASE + Number(__VU) * 1000000 + Number(__ITER); // 每次迭代新用户
    case 'PER_VU':
    default:
      return USER_ID_BASE + Number(__VU);                       // 每个 VU 固定一个用户
  }
}

function withClient(target, fn) {
  // 长连接：避免每次调用反复握手，降低端口耗尽风险
  if (!withClient._connected || withClient._target !== target) {
    try { client.close(); } catch (e) { /* 忽略关闭异常 */ }
    client.connect(target, { plaintext: true, timeout: '5s' }); // 建立新连接
    withClient._connected = true;
    withClient._target = target;
  }
  fn(); // 执行传入的函数体（发请求）
}

// 仅压测参与秒杀接口

// 每VU在一个长连接上按固定速率发起请求，实现更强连接复用
export function flashSaleLoop(data) {
  const perVuRps = Number(__ENV.PER_VU_RPS) || 10; // 每 VU 每秒请求数（目标）
  const interval = 1.0 / perVuRps;                // 期望的发送间隔（秒）
  const start = Date.now();                       // 记录本次迭代起始时间
  withClient(data.target, () => {
    const userId = pickUserId(); // 根据模式生成用户ID
    let res;                     // 保存 gRPC 响应
    try {
      res = client.invoke('Coupon/ParticipateFlashSale', {
        userId: userId,                 // 用户ID
        flashSaleId: FLASH_SALE_ID,     // 活动ID
      }, { timeout: __ENV.GRPC_TIMEOUT || '5s' }); // 调用超时（非连接超时）
    } catch (e) {
      flashFail.add(1);                                 // 统计失败
      check(null, { 'grpc ok': () => false });          // 记录一次失败的 check
      // 若请求异常，仍按节流目标补偿性休眠
      const used = (Date.now() - start) / 1000;
      const pad = interval - used;
      if (pad > 0) sleep(pad);
      return;
    }
    const ok = check(res, { 'grpc ok': (r) => r && r.status === grpc.StatusOK }); // gRPC 层面的 OK
    if (ok) {
      const msg = res.message || {};                    // 解析响应体
      if (msg.status === 1) flashSuccess.add(1);        // 业务成功（status==1）
      else flashFail.add(1);                            // 业务失败
    } else {
      flashFail.add(1);                                  // gRPC 非 StatusOK
    }
  });
  // 补偿性节流：仅在本次调用耗时小于目标间隔时休眠，避免“额外固定延迟”降低实际吞吐
  const used = (Date.now() - start) / 1000;
  const pad = interval - used;
  if (pad > 0) sleep(pad);
}

// CAR 模式：每次迭代只发一次请求，由执行器保证到达速率
export function flashSaleOnce(data) {
  withClient(data.target, () => {
    const userId = pickUserId();
    try {
      const res = client.invoke('Coupon/ParticipateFlashSale', {
        userId: userId,
        flashSaleId: FLASH_SALE_ID,
      }, { timeout: __ENV.GRPC_TIMEOUT || '5s' });
      const ok = check(res, { 'grpc ok': (r) => r && r.status === grpc.StatusOK });
      if (ok) {
        const msg = res.message || {};
        if (msg.status === 1) flashSuccess.add(1); else flashFail.add(1);
      } else {
        flashFail.add(1);
      }
    } catch (e) {
      flashFail.add(1);
      check(null, { 'grpc ok': () => false });
    }
  });
}

// 取消其它非必要接口的压测函数（库存查询/优惠计算），聚焦参与秒杀

// 自定义总结：输出 HTML 和 JSON（可通过环境变量关闭/改变路径）
export function handleSummary(data) {
  const out = {};
  const metrics = data.metrics || {};
  const get = (name, stat) => (metrics[name] && metrics[name].values && metrics[name].values[stat] !== undefined ? metrics[name].values[stat] : 'NA');
  const lines = [];
  lines.push('=== k6 Summary (basic) ===');
  lines.push(`iterations:  ${get('iterations','count')}`);
  lines.push(`vus (max):   ${get('vus','max')}`);
  lines.push(`grpc avg:    ${get('grpc_req_duration','avg')}  p95: ${get('grpc_req_duration','p(95)')}  max: ${get('grpc_req_duration','max')}`);
  lines.push(`success:     ${get('flash_sale_success_total','count')}  fail: ${get('flash_sale_fail_total','count')}`);
  lines.push('');
  lines.push('Config:');
  const vus = Number(__ENV.VUS) || 200;
  const perVu = Number(__ENV.PER_VU_RPS) || 10;
  const carRate = Number(__ENV.FLASH_PARTICIPATE_RPS) || 0;
  const useCVUSRate = String(__ENV.CVUS_MODE || '0') === '1';
  const rate = carRate > 0 ? carRate : (useCVUSRate ? vus * perVu : 0);
  const scenarioName = rate > 0 ? 'flashsale_rate' : 'flashsale_cvus';
  const cfg = rate > 0
    ? { scenario: scenarioName, rate, pre_vus: Number(__ENV.PRE_VUS) || vus, max_vus: Number(__ENV.MAX_VUS) || Math.max(Number(__ENV.PRE_VUS) || vus, vus), duration: __ENV.DURATION || '30s', timeout: __ENV.GRPC_TIMEOUT || '5s' }
    : { scenario: scenarioName, vus, per_vu_rps: perVu, duration: __ENV.DURATION || '30s', timeout: __ENV.GRPC_TIMEOUT || '5s' };
  lines.push(JSON.stringify(cfg, null, 2));
  out['stdout'] = lines.join('\n');

  const jsonPath = __ENV.SUMMARY_JSON || '';
  if (jsonPath) out[jsonPath] = JSON.stringify(data);
  const htmlPath = __ENV.SUMMARY_HTML || '';
  if (htmlPath) {
    const html = `<!doctype html><html><head><meta charset="utf-8"/>
<title>k6 Summary</title>
<style>body{font-family:ui-monospace,Menlo,Consolas,monospace;white-space:pre-wrap}</style>
</head><body>\n${out['stdout'].replace(/&/g,'&amp;').replace(/</g,'&lt;')}\n</body></html>`;
    out[htmlPath] = html;
  }
  return out;
}
