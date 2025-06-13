/**
 * Database Performance Focused Load Test
 * Tests database query performance, connection pooling, and optimization
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics for database performance
const dbQueryTime = new Trend('db_query_time');
const cacheHitRate = new Rate('cache_hit_rate');
const dbConnectionErrors = new Rate('db_connection_errors');
const complexQueryTime = new Trend('complex_query_time');
const paginationPerformance = new Trend('pagination_performance');

export const options = {
  stages: [
    { duration: '1m', target: 20 },    // Warm up database connections
    { duration: '5m', target: 80 },    // Sustained database load
    { duration: '2m', target: 120 },   // Peak database stress
    { duration: '3m', target: 80 },    // Return to sustained load
    { duration: '1m', target: 0 },     // Cool down
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],         // Database queries under 300ms
    db_query_time: ['p(95)<200'],             // Database-specific threshold
    cache_hit_rate: ['rate>0.6'],             // 60%+ cache hit rate
    db_connection_errors: ['rate<0.001'],     // Less than 0.1% connection errors
    complex_query_time: ['p(95)<500'],        // Complex queries under 500ms
    pagination_performance: ['p(95)<250'],    // Pagination queries under 250ms
  },
};

const BASE_URL = 'http://localhost:8080';
const API_KEY = 'api_key_pro_5678'; // Use pro tier for testing

// Database-intensive query patterns
const DB_TEST_PATTERNS = [
  {
    name: 'Simple_Pagination',
    weight: 25,
    endpoints: [
      '/api/news?page=1&limit=10',
      '/api/news?page=2&limit=10',
      '/api/news?page=3&limit=10',
    ],
    complexity: 'low'
  },
  {
    name: 'Category_Filtering',
    weight: 25,
    endpoints: [
      '/api/news?category=technology&page=1&limit=10',
      '/api/news?category=business&page=1&limit=15',
      '/api/news?category=sports&page=2&limit=10',
    ],
    complexity: 'medium'
  },
  {
    name: 'Complex_Filtering',
    weight: 20,
    endpoints: [
      '/api/news?category=technology&q=AI&page=1&limit=10',
      '/api/news?q=market&page=1&limit=20',
      '/api/news?category=world&q=economy&page=2&limit=15',
    ],
    complexity: 'high'
  },
  {
    name: 'Large_Pagination',
    weight: 15,
    endpoints: [
      '/api/news?page=1&limit=50',
      '/api/news?page=5&limit=20',
      '/api/news?page=10&limit=10',
    ],
    complexity: 'medium'
  },
  {
    name: 'Deep_Pagination',
    weight: 10,
    endpoints: [
      '/api/news?page=20&limit=10',
      '/api/news?page=50&limit=5',
      '/api/news?page=100&limit=5',
    ],
    complexity: 'high'
  },
  {
    name: 'Analytics_Queries',
    weight: 5,
    endpoints: [
      '/api/analytics',
      '/api/analytics?timeframe=week',
      '/api/analytics?category=technology',
    ],
    complexity: 'very_high'
  }
];

function selectDbTestPattern() {
  const random = Math.random() * 100;
  let weightSum = 0;
  
  for (const pattern of DB_TEST_PATTERNS) {
    weightSum += pattern.weight;
    if (random <= weightSum) {
      return pattern;
    }
  }
  return DB_TEST_PATTERNS[0];
}

export default function () {
  const pattern = selectDbTestPattern();
  const endpoint = pattern.endpoints[Math.floor(Math.random() * pattern.endpoints.length)];
  const url = `${BASE_URL}${endpoint}`;
  
  const startTime = Date.now();
  const response = http.get(url, {
    headers: { 'X-API-Key': API_KEY },
    tags: {
      pattern: pattern.name,
      complexity: pattern.complexity,
      test_type: 'database_performance'
    }
  });
  
  const responseTime = response.timings.duration;
  const totalTime = Date.now() - startTime;
  
  // Record database-specific metrics
  dbQueryTime.add(responseTime);
  
  // Categorize performance by complexity
  switch (pattern.complexity) {
    case 'very_high':
      complexQueryTime.add(responseTime);
      break;
    case 'high':
      complexQueryTime.add(responseTime);
      break;
    case 'medium':
      if (pattern.name.includes('Pagination')) {
        paginationPerformance.add(responseTime);
      }
      break;
    case 'low':
      paginationPerformance.add(responseTime);
      break;
  }
  
  // Detect cache hits (fast responses likely from cache)
  const isCacheHit = responseTime < 50 || 
    (response.headers['Cache-Control'] && responseTime < 100);
  cacheHitRate.add(isCacheHit);
  
  // Detect database connection issues
  const hasDbError = response.status === 500 || 
    response.status === 503 || 
    response.status === 0;
  dbConnectionErrors.add(hasDbError);
  
  // Comprehensive database performance checks
  check(response, {
    'database query successful': (r) => r.status === 200,
    'response time reasonable for complexity': (r) => {
      const limits = {
        'low': 100,
        'medium': 200,
        'high': 400,
        'very_high': 800
      };
      return r.timings.duration < limits[pattern.complexity];
    },
    'no database connection errors': (r) => r.status !== 503,
    'response has valid pagination': (r) => {
      if (r.status !== 200) return true;
      try {
        const body = JSON.parse(r.body);
        return body.page !== undefined && body.totalItems !== undefined;
      } catch {
        return false;
      }
    },
    'data consistency check': (r) => {
      if (r.status !== 200) return true;
      try {
        const body = JSON.parse(r.body);
        return Array.isArray(body.data) && body.data.length <= body.limit;
      } catch {
        return false;
      }
    },
    'cache headers present': (r) => 
      r.headers['Cache-Control'] !== undefined || r.headers['ETag'] !== undefined
  }, {
    pattern: pattern.name,
    complexity: pattern.complexity
  });
  
  // Query-specific think time
  const thinkTime = {
    'low': 0.5,
    'medium': 1,
    'high': 1.5,
    'very_high': 2
  }[pattern.complexity] || 1;
  
  sleep(Math.random() * thinkTime + 0.5);
}

export function handleSummary(data) {
  return {
    'reports/database-performance-summary.json': JSON.stringify(data, null, 2),
    stdout: createDbSummary(data),
  };
}

function createDbSummary(data) {
  const summary = `
🗄️  Database Performance Load Test Summary
=========================================

📊 Database Query Performance:
- Total Database Queries: ${data.metrics.http_reqs.count}
- Average Query Time: ${data.metrics.db_query_time.avg.toFixed(2)}ms
- 95th Percentile Query Time: ${data.metrics.db_query_time['p(95)'].toFixed(2)}ms
- Complex Query P95: ${(data.metrics.complex_query_time?.['p(95)'] || 0).toFixed(2)}ms
- Pagination Query P95: ${(data.metrics.pagination_performance?.['p(95)'] || 0).toFixed(2)}ms

🚀 Cache Performance:
- Cache Hit Rate: ${((data.metrics.cache_hit_rate?.rate || 0) * 100).toFixed(2)}%
- Cache Effectiveness: ${getCacheEffectiveness(data)}

🔗 Connection Health:
- Database Connection Errors: ${((data.metrics.db_connection_errors?.rate || 0) * 100).toFixed(4)}%
- Connection Stability: ${getConnectionStability(data)}

⚡ Query Complexity Analysis:
${getComplexityAnalysis(data)}

✅ Database Performance Status: ${getDbPerformanceStatus(data)}

🔧 Database Optimization Recommendations:
${getDbRecommendations(data)}
`;
  return summary;
}

function getCacheEffectiveness(data) {
  const cacheRate = data.metrics.cache_hit_rate?.rate || 0;
  if (cacheRate > 0.8) return '🟢 EXCELLENT';
  if (cacheRate > 0.6) return '🟡 GOOD';
  if (cacheRate > 0.4) return '🟠 MODERATE';
  return '🔴 POOR';
}

function getConnectionStability(data) {
  const errorRate = data.metrics.db_connection_errors?.rate || 0;
  if (errorRate < 0.001) return '🟢 STABLE';
  if (errorRate < 0.005) return '🟡 MOSTLY STABLE';
  return '🔴 UNSTABLE';
}

function getComplexityAnalysis(data) {
  const dbP95 = data.metrics.db_query_time?.['p(95)'] || 0;
  const complexP95 = data.metrics.complex_query_time?.['p(95)'] || 0;
  const paginationP95 = data.metrics.pagination_performance?.['p(95)'] || 0;
  
  return `• Simple Queries (Pagination): ${paginationP95.toFixed(2)}ms P95
• Complex Queries (Filtering): ${complexP95.toFixed(2)}ms P95
• Overall Database Performance: ${dbP95.toFixed(2)}ms P95`;
}

function getDbPerformanceStatus(data) {
  const dbP95 = data.metrics.db_query_time?.['p(95)'] || 0;
  const cacheRate = data.metrics.cache_hit_rate?.rate || 0;
  const errorRate = data.metrics.db_connection_errors?.rate || 0;
  
  if (dbP95 < 200 && cacheRate > 0.7 && errorRate < 0.001) return '🟢 EXCELLENT DB PERFORMANCE';
  if (dbP95 < 300 && cacheRate > 0.6 && errorRate < 0.005) return '🟡 GOOD DB PERFORMANCE';
  if (dbP95 < 500 && errorRate < 0.01) return '🟠 ACCEPTABLE DB PERFORMANCE';
  return '🔴 POOR DB PERFORMANCE - IMMEDIATE OPTIMIZATION NEEDED';
}

function getDbRecommendations(data) {
  const recommendations = [];
  const dbP95 = data.metrics.db_query_time?.['p(95)'] || 0;
  const complexP95 = data.metrics.complex_query_time?.['p(95)'] || 0;
  const cacheRate = data.metrics.cache_hit_rate?.rate || 0;
  const errorRate = data.metrics.db_connection_errors?.rate || 0;
  
  if (dbP95 > 300) recommendations.push('• Add database indexes for frequently queried columns');
  if (complexP95 > 500) recommendations.push('• Optimize complex JOIN queries and consider query restructuring');
  if (cacheRate < 0.6) recommendations.push('• Improve caching strategy - increase TTL for stable data');
  if (errorRate > 0.001) recommendations.push('• Investigate connection pool configuration and database stability');
  
  const paginationP95 = data.metrics.pagination_performance?.['p(95)'] || 0;
  if (paginationP95 > 250) recommendations.push('• Optimize pagination queries - consider cursor-based pagination');
  
  if (dbP95 > 200) recommendations.push('• Consider read replicas for query distribution');
  if (recommendations.length === 0) {
    recommendations.push('• Database performance is optimal!');
  }
  
  return recommendations.join('\n');
}
