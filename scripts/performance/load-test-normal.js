/**
 * Normal User Traffic Load Test
 * Simulates typical browsing patterns with realistic user behavior
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('error_rate');
const responseTimeP95 = new Trend('response_time_p95');

export const options = {
  stages: [
    { duration: '2m', target: 10 },   // Warm up
    { duration: '5m', target: 50 },   // Normal load
    { duration: '10m', target: 50 },  // Sustained normal load
    { duration: '2m', target: 0 },    // Cool down
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'],     // 95% of requests should be below 200ms
    http_req_failed: ['rate<0.01'],       // Error rate should be less than 1%
    error_rate: ['rate<0.01'],            // Custom error rate should be less than 1%
    http_reqs: ['rate>100'],              // Should handle more than 100 RPS
  },
};

const BASE_URL = 'http://localhost:8080';

// API keys for different tiers
const API_KEYS = {
  basic: 'api_key_basic_1234',
  pro: 'api_key_pro_5678',
  enterprise: 'api_key_enterprise_9012'
};

// Realistic browsing patterns
const BROWSING_PATTERNS = [
  {
    name: 'Homepage_Browse',
    weight: 40,
    requests: [
      { endpoint: '/api/news', params: '?page=1&limit=10' },
      { endpoint: '/api/news', params: '?page=2&limit=10' },
    ]
  },
  {
    name: 'Category_Browse',
    weight: 30,
    requests: [
      { endpoint: '/api/news', params: '?category=technology&page=1&limit=10' },
      { endpoint: '/api/news', params: '?category=business&page=1&limit=10' },
      { endpoint: '/api/news', params: '?category=sports&page=1&limit=10' },
    ]
  },
  {
    name: 'Search_Pattern',
    weight: 20,
    requests: [
      { endpoint: '/api/news', params: '?q=breaking&page=1&limit=10' },
      { endpoint: '/api/news', params: '?q=latest&page=1&limit=10' },
    ]
  },
  {
    name: 'Deep_Browse',
    weight: 10,
    requests: [
      { endpoint: '/api/news', params: '?page=1&limit=20' },
      { endpoint: '/api/news', params: '?page=3&limit=10' },
      { endpoint: '/api/news', params: '?page=5&limit=10' },
    ]
  }
];

function selectRandomPattern() {
  const random = Math.random() * 100;
  let weightSum = 0;
  
  for (const pattern of BROWSING_PATTERNS) {
    weightSum += pattern.weight;
    if (random <= weightSum) {
      return pattern;
    }
  }
  return BROWSING_PATTERNS[0]; // fallback
}

function selectRandomApiKey() {
  const keys = Object.values(API_KEYS);
  return keys[Math.floor(Math.random() * keys.length)];
}

export default function () {
  const pattern = selectRandomPattern();
  const apiKey = selectRandomApiKey();
  
  // Simulate user session
  for (const request of pattern.requests) {
    const url = `${BASE_URL}${request.endpoint}${request.params}`;
    const headers = { 'X-API-Key': apiKey };
    
    const response = http.get(url, {
      headers,
      tags: {
        pattern: pattern.name,
        endpoint: request.endpoint,
        tier: getTierFromApiKey(apiKey)
      }
    });
    
    // Record metrics
    errorRate.add(response.status !== 200);
    responseTimeP95.add(response.timings.duration);
    
    // Validate response
    check(response, {
      'status is 200': (r) => r.status === 200,
      'response time < 200ms': (r) => r.timings.duration < 200,
      'response has data': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.data && Array.isArray(body.data);
        } catch {
          return false;
        }
      },
      'cache headers present': (r) => 
        r.headers['Cache-Control'] !== undefined,
      'pagination metadata present': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.page !== undefined && body.totalItems !== undefined;
        } catch {
          return false;
        }
      }
    }, {
      pattern: pattern.name,
      endpoint: request.endpoint
    });
    
    // Realistic user think time
    sleep(Math.random() * 2 + 1); // 1-3 seconds
  }
  
  // Session break
  sleep(Math.random() * 5 + 2); // 2-7 seconds between sessions
}

function getTierFromApiKey(apiKey) {
  if (apiKey.includes('basic')) return 'basic';
  if (apiKey.includes('pro')) return 'pro';
  if (apiKey.includes('enterprise')) return 'enterprise';
  return 'unknown';
}

export function handleSummary(data) {
  return {
    'reports/normal-traffic-summary.json': JSON.stringify(data, null, 2),
    stdout: createTextSummary(data),
  };
}

function createTextSummary(data) {
  const summary = `
📊 Normal Traffic Load Test Summary
================================

🎯 Test Results:
- Total Requests: ${data.metrics.http_reqs.count}
- Request Rate: ${data.metrics.http_reqs.rate.toFixed(2)} RPS
- Error Rate: ${(data.metrics.http_req_failed.rate * 100).toFixed(2)}%

⏱️  Response Times:
- Average: ${data.metrics.http_req_duration.avg.toFixed(2)}ms
- 95th Percentile: ${data.metrics.http_req_duration['p(95)'].toFixed(2)}ms
- 99th Percentile: ${data.metrics.http_req_duration['p(99)'].toFixed(2)}ms

✅ Thresholds:
- Response Time P95 < 200ms: ${data.metrics.http_req_duration['p(95)'] < 200 ? 'PASS' : 'FAIL'}
- Error Rate < 1%: ${data.metrics.http_req_failed.rate < 0.01 ? 'PASS' : 'FAIL'}
- Request Rate > 100 RPS: ${data.metrics.http_reqs.rate > 100 ? 'PASS' : 'FAIL'}

📈 Performance Status: ${getOverallStatus(data)}
`;
  return summary;
}

function getOverallStatus(data) {
  const p95Pass = data.metrics.http_req_duration['p(95)'] < 200;
  const errorPass = data.metrics.http_req_failed.rate < 0.01;
  const rpsPass = data.metrics.http_reqs.rate > 100;
  
  if (p95Pass && errorPass && rpsPass) return '🟢 EXCELLENT';
  if (p95Pass && errorPass) return '🟡 GOOD';
  return '🔴 NEEDS IMPROVEMENT';
}
