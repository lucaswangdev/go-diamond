import http from 'k6/http';
import { check, sleep } from 'k6';

// Test configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test options - simulate varying load
export const options = {
  stages: [
    { duration: '10s', target: 10 },    // Light load
    { duration: '30s', target: 50 },   // Medium load
    { duration: '1m', target: 100 },   // Heavy load
    { duration: '30s', target: 0 },     // Cool down
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],  // 95% under 1s
    http_req_failed: ['rate<0.1'],       // Less than 10% failures
  },
};

export default function () {
  // Test 1: Health check endpoint
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health check status 200': (r) => r.status === 200,
    'health check response time < 100ms': (r) => r.timings.duration < 100,
  });

  // Test 2: Ready endpoint
  const readyRes = http.get(`${BASE_URL}/ready`);
  check(readyRes, {
    'ready status 200': (r) => r.status === 200,
  });

  // Test 3: Get config (public endpoint)
  const namespace = 'default';
  const group = 'DEFAULT_GROUP';
  const dataId = 'app.json';

  const configRes = http.get(`${BASE_URL}/api/v1/configs/${namespace}/${group}/${dataId}`);
  check(configRes, {
    'config endpoint accessible': (r) => r.status === 200 || r.status === 404,
    'config response time < 200ms': (r) => r.timings.duration < 200,
  });

  // Test 4: Watch endpoint (long polling)
  const watchRes = http.get(`${BASE_URL}/api/v1/watch/${namespace}/${group}/${dataId}?timeout=5`);
  check(watchRes, {
    'watch endpoint accessible': (r) => r.status === 200 || r.status === 304 || r.status === 404,
    'watch response time < 10s': (r) => r.timings.duration < 10000,
  });

  // Test 5: Batch watch endpoint
  const batchPayload = JSON.stringify({
    watchItems: [
      { namespace, group, dataId, md5: '' },
      { namespace, group, dataId: 'db.json', md5: '' },
    ],
    timeout: 5,
  });

  const batchRes = http.post(
    `${BASE_URL}/api/v1/watch/batch`,
    batchPayload,
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(batchRes, {
    'batch watch accessible': (r) => r.status === 200 || r.status === 404,
    'batch response time < 500ms': (r) => r.timings.duration < 500,
  });

  // Simulate some delay between requests
  sleep(0.1);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'summary.json': JSON.stringify(data, null, 2),
  };
}

function textSummary(data, opts) {
  const indent = opts.indent || '';
  let summary = `${indent}API Load Test Summary\n`;
  summary += `${indent}======================\n\n`;
  summary += `${indent}Total Requests: ${data.metrics.http_reqs.values.count}\n`;
  summary += `${indent}Request Rate: ${data.metrics.http_reqs.values.rate.toFixed(2)}/s\n`;
  summary += `${indent}Avg Duration: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
  summary += `${indent}p95 Duration: ${data.metrics.http_req_duration.values['p(95)']?.toFixed(2) ?? 0}ms\n`;
  summary += `${indent}p99 Duration: ${data.metrics.http_req_duration.values['p(99)']?.toFixed(2) ?? 0}ms\n`;
  summary += `${indent}Max Duration: ${data.metrics.http_req_duration.values.max?.toFixed(2) ?? 0}ms\n`;
  summary += `${indent}Failed Rate: ${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%\n`;
  summary += `${indent}Request Duration - Avg: ${data.metrics.http_req_duration.values.avg}ms\n`;
  return summary;
}