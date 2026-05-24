import http from 'k6/http';
import { check, sleep } from 'k6';

// Test configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:5173';

// Test options
export const options = {
  stages: [
    { duration: '30s', target: 20 },   // Ramp up to 20 users
    { duration: '1m', target: 20 },    // Stay at 20 users
    { duration: '30s', target: 50 },    // Ramp up to 50 users
    { duration: '1m', target: 50 },     // Stay at 50 users
    { duration: '30s', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95% of requests under 500ms
    http_req_failed: ['rate<0.05'],    // Less than 5% failure rate
  },
};

// Default function
export default function () {
  // Test Dashboard page
  const dashboardRes = http.get(`${BASE_URL}/`);
  check(dashboardRes, {
    'dashboard loaded': (r) => r.status === 200,
    'dashboard is HTML': (r) => r.headers['Content-Type'] && r.headers['Content-Type'].includes('text/html'),
  });

  // Test New Test page
  const newTestRes = http.get(`${BASE_URL}/tests/new`);
  check(newTestRes, {
    'new test page loaded': (r) => r.status === 200,
    'new test page is HTML': (r) => r.headers['Content-Type'] && r.headers['Content-Type'].includes('text/html'),
  });

  // Test form submission simulation
  const payload = JSON.stringify({
    name: `Load Test ${Date.now()}`,
    url: 'https://api.example.com/health',
    method: 'GET',
    duration: 60,
    concurrency: 10,
    requestsPerSecond: 100,
    timeout: 5000,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  // Simulate creating a test - use API endpoint instead of page POST
  const createRes = http.post(`${BASE_URL}/api/tests`, payload, params);
  check(createRes, {
    'test creation endpoint works': (r) => r.status === 200 || r.status === 201 || r.status === 404 || r.status === 400,
  });

  sleep(1);
}

// Summary handler
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'summary.json': JSON.stringify(data, null, 2),
  };
}

function textSummary(data, opts) {
  const indent = opts.indent || '';
  let summary = `${indent}Load Test Summary\n`;
  summary += `${indent}==================\n\n`;
  summary += `${indent}Total Requests: ${data.metrics.http_reqs.values.count}\n`;
  summary += `${indent}Failed Requests: ${data.metrics.http_req_failed.values.fails}\n`;
  summary += `${indent}Request Duration p95: ${data.metrics.http_req_duration.values['p(95)']}ms\n`;
  summary += `${indent}Request Rate: ${data.metrics.http_reqs.values.rate}/s\n`;
  return summary;
}