/**
 * High Traffic Spike Load Test
 * Simulates sudden traffic spikes during breaking news events
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('spike_error_rate');
const timeoutRate = new Rate('timeout_rate');
const cacheHitRate = new Rate('cache_hit_rate');
const dbConnections = new Trend('db_connection_time');
const concurrentUsers = new Counter('concurrent_users');

export const options = {
  stages: [
    { duration: '30s', target: 50 },    // Normal baseline
    { duration: '1m', target: 300 },    // Rapid spike to 300 users
    { duration: '2m', target: 500 },    // Peak traffic (breaking news)
    { duration: '1m', target: 300 },    // Gradual decline
    { duration: '30s', target: 50 },    // Return to normal
    { duration: '30s', target: 0 },     // Cool down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],    // Relaxed threshold for spike
    http_req_failed: ['rate<0.05'],      // Allow up to 5% error rate during spike
    spike_error_rate: ['rate<0.05'],     // Custom spike error tracking
    timeout_rate: ['rate<0.02'],         // Less than 2% timeouts
    http_reqs: ['rate>200'],             // Should handle more than 200 RPS
  },
};

const BASE_URL = 'http://localhost:8080';

// Breaking news simulation endpoints
const BREAKING_NEWS_PATTERNS = [
  {
    name: 'Latest_Breaking',
    weight: 50,
    endpoints: [
      '/api/news?category=breaking&page=1&limit=10',
      '/api/news?q=breaking&page=1&limit=20',
      '/api/news?page=1&limit=15',
    ]
  },
  {
    name: 'Category_Rush',
    weight: 30,
    endpoints: [
      '/api/news?category=politics&page=1&limit=10',
      '/api/news?category=world&page=1&limit=10',
      '/api/news?category=technology&page=1&limit=10',
    ]
  },
  {
    name: 'Deep_Pagination',
    weight: 20,
    endpoints: [
      '/api/news?page=2&limit=10',
      '/api/news?page=3&limit=10',
      '/api/news?page=1&limit=50',
    ]
  }
];

const API_KEYS = [
  'api_key_basic_1234',
  'api_key_pro_5678',
  'api_key_enterprise_9012'
];

function selectBreakingNewsPattern() {
  const random = Math.random() * 100;
  let weightSum = 0;
  
  for (const pattern of BREAKING_NEWS_PATTERNS) {
    weightSum += pattern.weight;
    if (random <= weightSum) {
      return pattern;
    }
  }
  return BREAKING_NEWS_PATTERNS[0];
}

export default function () {
  const pattern = selectBreakingNewsPattern();
  const apiKey = API_KEYS[Math.floor(Math.random() * API_KEYS.length)];
  
  concurrentUsers.add(1);
  
  // Simulate urgent browsing behavior during breaking news
  const endpoint = pattern.endpoints[Math.floor(Math.random() * pattern.endpoints.length)];
  const url = `${BASE_URL}${endpoint}`;
  
  const startTime = Date.now();
  const response = http.get(url, {
    headers: { 'X-API-Key': apiKey },
    timeout: '10s', // 10 second timeout
    tags: {
      pattern: pattern.name,
      endpoint: endpoint,
      scenario: 'breaking_news'
    }
  });
  
  const responseTime = Date.now() - startTime;
  
  // Record custom metrics
  errorRate.add(response.status !== 200);
  timeoutRate.add(response.status === 0); // Timeout
  
  // Check for cache hits (faster responses likely from cache)
  const isCacheHit = responseTime < 100 || 
    response.headers['X-Cache-Status'] === 'HIT';
  cacheHitRate.add(isCacheHit);
  
  // Validate spike performance
  check(response, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'response time acceptable for spike': (r) => r.timings.duration < 1000,
    'not timeout': (r) => r.status !== 0,
    'has rate limit headers during spike': (r) => 
      r.status === 429 ? r.headers['Retry-After'] !== undefined : true,
    'response structure valid': (r) => {
      if (r.status !== 200) return true; // Skip validation for non-200
      try {
        const body = JSON.parse(r.body);
        return body.data !== undefined;
      } catch {
        return false;
      }
    }
  }, {
    pattern: pattern.name,
    scenario: 'breaking_news'
  });
  
  // Aggressive browsing during breaking news - minimal think time
  if (response.status === 200) {
    sleep(Math.random() * 0.5 + 0.1); // 0.1-0.6 seconds (very fast)
  } else if (response.status === 429) {
    // Respect rate limiting
    const retryAfter = parseInt(response.headers['Retry-After']) || 1;
    sleep(retryAfter);
  } else {
    sleep(1); // Error recovery time
  }
}

export function handleSummary(data) {
  return {
    'reports/spike-traffic-summary.json': JSON.stringify(data, null, 2),
    stdout: createSpikeSummary(data),
  };
}

function createSpikeSummary(data) {
  const summary = `
🚨 Traffic Spike Load Test Summary
=================================

🎯 Peak Performance Results:
- Total Requests: ${data.metrics.http_reqs.count}
- Peak Request Rate: ${data.metrics.http_reqs.rate.toFixed(2)} RPS
- Error Rate During Spike: ${(data.metrics.http_req_failed.rate * 100).toFixed(2)}%
- Timeout Rate: ${((data.metrics.timeout_rate?.rate || 0) * 100).toFixed(2)}%

⚡ Response Time Performance:
- Average: ${data.metrics.http_req_duration.avg.toFixed(2)}ms
- 95th Percentile: ${data.metrics.http_req_duration['p(95)'].toFixed(2)}ms
- 99th Percentile: ${data.metrics.http_req_duration['p(99)'].toFixed(2)}ms
- Max Response Time: ${data.metrics.http_req_duration.max.toFixed(2)}ms

📊 Spike Resilience:
- Cache Hit Rate: ${((data.metrics.cache_hit_rate?.rate || 0) * 100).toFixed(2)}%
- Rate Limiting Effectiveness: ${getRateLimitingEffectiveness(data)}
- System Stability: ${getSystemStability(data)}

✅ Spike Performance Status: ${getSpikeStatus(data)}

🔧 Recommendations:
${getRecommendations(data)}
`;
  return summary;
}

function getRateLimitingEffectiveness(data) {
  const errorRate = data.metrics.http_req_failed.rate;
  if (errorRate < 0.02) return '🟢 EXCELLENT';
  if (errorRate < 0.05) return '🟡 GOOD';
  return '🔴 NEEDS TUNING';
}

function getSystemStability(data) {
  const timeoutRate = data.metrics.timeout_rate?.rate || 0;
  if (timeoutRate < 0.01) return '🟢 STABLE';
  if (timeoutRate < 0.02) return '🟡 MOSTLY STABLE';
  return '🔴 UNSTABLE';
}

function getSpikeStatus(data) {
  const p95 = data.metrics.http_req_duration['p(95)'];
  const errorRate = data.metrics.http_req_failed.rate;
  const timeoutRate = data.metrics.timeout_rate?.rate || 0;
  
  if (p95 < 500 && errorRate < 0.02 && timeoutRate < 0.01) return '🟢 EXCELLENT SPIKE HANDLING';
  if (p95 < 800 && errorRate < 0.05 && timeoutRate < 0.02) return '🟡 GOOD SPIKE RESILIENCE';
  return '🔴 POOR SPIKE PERFORMANCE - OPTIMIZATION NEEDED';
}

function getRecommendations(data) {
  const recommendations = [];
  const p95 = data.metrics.http_req_duration['p(95)'];
  const errorRate = data.metrics.http_req_failed.rate;
  const timeoutRate = data.metrics.timeout_rate?.rate || 0;
  
  if (p95 > 500) recommendations.push('• Optimize response times - consider caching improvements');
  if (errorRate > 0.03) recommendations.push('• Review rate limiting configuration');
  if (timeoutRate > 0.01) recommendations.push('• Investigate timeout causes - database/external services');
  
  const cacheHitRate = data.metrics.cache_hit_rate?.rate || 0;
  if (cacheHitRate < 0.5) recommendations.push('• Improve cache strategy for breaking news scenarios');
  
  if (recommendations.length === 0) {
    recommendations.push('• System performs well under spike conditions!');
  }
  
  return recommendations.join('\n');
}
