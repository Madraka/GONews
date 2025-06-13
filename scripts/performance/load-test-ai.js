/**
 * AI Integration Performance Load Test
 * Tests performance of AI processing endpoints and external API integrations
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics for AI performance
const aiProcessingTime = new Trend('ai_processing_time');
const aiErrorRate = new Rate('ai_error_rate');
const externalApiTime = new Trend('external_api_time');
const aiQueueDepth = new Counter('ai_queue_depth');
const successfulAiProcessing = new Rate('successful_ai_processing');

export const options = {
  stages: [
    { duration: '1m', target: 5 },     // Warm up AI services
    { duration: '3m', target: 15 },    // Moderate AI load
    { duration: '2m', target: 25 },    // Peak AI processing
    { duration: '2m', target: 15 },    // Sustained AI load
    { duration: '1m', target: 0 },     // Cool down
  ],
  thresholds: {
    http_req_duration: ['p(95)<5000'],        // AI requests can be slower
    ai_processing_time: ['p(95)<3000'],       // AI processing under 3s
    ai_error_rate: ['rate<0.05'],             // 5% AI error tolerance
    external_api_time: ['p(95)<2000'],        // External API under 2s
    successful_ai_processing: ['rate>0.9'],   // 90% success rate
  },
};

const BASE_URL = 'http://localhost:8080';

// AI test scenarios with different complexity levels
const AI_TEST_SCENARIOS = [
  {
    name: 'Text_Summarization',
    weight: 40,
    endpoints: [
      '/api/ai/summarize',
    ],
    payload: {
      text: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.',
      maxLength: 100
    },
    complexity: 'medium',
    expectedDuration: 2000
  },
  {
    name: 'Content_Analysis',
    weight: 30,
    endpoints: [
      '/api/ai/analyze',
    ],
    payload: {
      content: 'Technology stocks surged today as investors showed confidence in the artificial intelligence sector. Major tech companies reported strong earnings.',
      analysisType: 'sentiment'
    },
    complexity: 'low',
    expectedDuration: 1500
  },
  {
    name: 'Category_Classification',
    weight: 20,
    endpoints: [
      '/api/ai/classify',
    ],
    payload: {
      title: 'Breaking: New AI Model Achieves Record Performance',
      content: 'Researchers have developed a new artificial intelligence model that outperforms previous benchmarks in natural language processing tasks.',
    },
    complexity: 'low',
    expectedDuration: 1000
  },
  {
    name: 'Complex_Processing',
    weight: 10,
    endpoints: [
      '/api/ai/process',
    ],
    payload: {
      text: 'This is a complex text that requires advanced natural language processing including entity recognition, sentiment analysis, and topic modeling.',
      features: ['entities', 'sentiment', 'topics', 'keywords']
    },
    complexity: 'high',
    expectedDuration: 4000
  }
];

// API keys with AI processing capabilities
const AI_API_KEYS = [
  'api_key_pro_5678',         // Pro tier with AI access
  'api_key_enterprise_9012',  // Enterprise tier with unlimited AI
];

function selectAiScenario() {
  const random = Math.random() * 100;
  let weightSum = 0;
  
  for (const scenario of AI_TEST_SCENARIOS) {
    weightSum += scenario.weight;
    if (random <= weightSum) {
      return scenario;
    }
  }
  return AI_TEST_SCENARIOS[0];
}

export default function () {
  const scenario = selectAiScenario();
  const apiKey = AI_API_KEYS[Math.floor(Math.random() * AI_API_KEYS.length)];
  const endpoint = scenario.endpoints[0];
  const url = `${BASE_URL}${endpoint}`;
  
  aiQueueDepth.add(1);
  
  const startTime = Date.now();
  const response = http.post(url, JSON.stringify(scenario.payload), {
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': apiKey
    },
    timeout: '10s', // Longer timeout for AI processing
    tags: {
      scenario: scenario.name,
      complexity: scenario.complexity,
      test_type: 'ai_performance'
    }
  });
  
  const processingTime = response.timings.duration;
  const totalTime = Date.now() - startTime;
  
  // Record AI-specific metrics
  aiProcessingTime.add(processingTime);
  externalApiTime.add(processingTime);
  
  const isSuccess = response.status === 200 || response.status === 202;
  aiErrorRate.add(!isSuccess);
  successfulAiProcessing.add(isSuccess);
  
  // AI endpoint validation
  check(response, {
    'AI request successful': (r) => r.status === 200 || r.status === 202,
    'processing time reasonable for complexity': (r) => {
      return r.timings.duration < scenario.expectedDuration;
    },
    'no AI service timeout': (r) => r.status !== 0,
    'no AI service overload': (r) => r.status !== 503,
    'valid AI response structure': (r) => {
      if (r.status !== 200) return true; // Skip validation for non-200
      try {
        const body = JSON.parse(r.body);
        return body.result !== undefined || body.data !== undefined;
      } catch {
        return false;
      }
    },
    'AI processing headers present': (r) => 
      r.headers['X-Processing-Time'] !== undefined || 
      r.headers['X-AI-Model'] !== undefined,
    'reasonable AI response size': (r) => {
      return r.body.length > 10 && r.body.length < 10000; // Reasonable response size
    }
  }, {
    scenario: scenario.name,
    complexity: scenario.complexity
  });
  
  // Handle different response scenarios
  if (response.status === 202) {
    // Async processing - simulate polling
    sleep(2);
    
    // Poll for result (simplified)
    const pollResponse = http.get(`${url}/status`, {
      headers: { 'X-API-Key': apiKey },
      tags: { scenario: scenario.name, operation: 'poll' }
    });
    
    check(pollResponse, {
      'polling successful': (r) => r.status === 200,
      'processing status available': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.status !== undefined;
        } catch {
          return false;
        }
      }
    });
  }
  
  // AI processing think time (simulates user waiting for AI results)
  const thinkTime = {
    'low': 2,
    'medium': 3,
    'high': 5
  }[scenario.complexity] || 3;
  
  sleep(Math.random() * thinkTime + 1);
}

export function handleSummary(data) {
  return {
    'reports/ai-performance-summary.json': JSON.stringify(data, null, 2),
    stdout: createAiSummary(data),
  };
}

function createAiSummary(data) {
  const summary = `
🤖 AI Integration Performance Test Summary
=========================================

🧠 AI Processing Performance:
- Total AI Requests: ${data.metrics.http_reqs.count}
- Average AI Processing Time: ${data.metrics.ai_processing_time.avg.toFixed(2)}ms
- 95th Percentile AI Time: ${data.metrics.ai_processing_time['p(95)'].toFixed(2)}ms
- Maximum AI Processing Time: ${data.metrics.ai_processing_time.max.toFixed(2)}ms

✅ AI Success Metrics:
- AI Success Rate: ${((data.metrics.successful_ai_processing?.rate || 0) * 100).toFixed(2)}%
- AI Error Rate: ${((data.metrics.ai_error_rate?.rate || 0) * 100).toFixed(2)}%
- External API Performance: ${(data.metrics.external_api_time?.['p(95)'] || 0).toFixed(2)}ms P95

🔄 Processing Analysis:
- AI Queue Efficiency: ${getQueueEfficiency(data)}
- Processing Stability: ${getProcessingStability(data)}
- Model Performance: ${getModelPerformance(data)}

📊 AI Service Health:
${getAiServiceHealth(data)}

✅ AI Integration Status: ${getAiIntegrationStatus(data)}

🔧 AI Optimization Recommendations:
${getAiRecommendations(data)}
`;
  return summary;
}

function getQueueEfficiency(data) {
  const avgTime = data.metrics.ai_processing_time?.avg || 0;
  if (avgTime < 1500) return '🟢 EXCELLENT';
  if (avgTime < 2500) return '🟡 GOOD';
  return '🔴 SLOW';
}

function getProcessingStability(data) {
  const errorRate = data.metrics.ai_error_rate?.rate || 0;
  const successRate = data.metrics.successful_ai_processing?.rate || 0;
  
  if (successRate > 0.95 && errorRate < 0.02) return '🟢 VERY STABLE';
  if (successRate > 0.9 && errorRate < 0.05) return '🟡 STABLE';
  return '🔴 UNSTABLE';
}

function getModelPerformance(data) {
  const p95 = data.metrics.ai_processing_time?.['p(95)'] || 0;
  if (p95 < 2000) return '🟢 FAST INFERENCE';
  if (p95 < 3000) return '🟡 MODERATE SPEED';
  return '🔴 SLOW INFERENCE';
}

function getAiServiceHealth(data) {
  const successRate = data.metrics.successful_ai_processing?.rate || 0;
  const avgTime = data.metrics.ai_processing_time?.avg || 0;
  
  return `• Success Rate: ${(successRate * 100).toFixed(1)}% ${successRate > 0.9 ? '✅' : '❌'}
• Average Response: ${avgTime.toFixed(0)}ms ${avgTime < 2000 ? '✅' : '❌'}
• Service Availability: ${successRate > 0.95 ? 'HIGH' : successRate > 0.9 ? 'MEDIUM' : 'LOW'}`;
}

function getAiIntegrationStatus(data) {
  const successRate = data.metrics.successful_ai_processing?.rate || 0;
  const p95 = data.metrics.ai_processing_time?.['p(95)'] || 0;
  const errorRate = data.metrics.ai_error_rate?.rate || 0;
  
  if (successRate > 0.95 && p95 < 3000 && errorRate < 0.02) return '🟢 EXCELLENT AI INTEGRATION';
  if (successRate > 0.9 && p95 < 4000 && errorRate < 0.05) return '🟡 GOOD AI INTEGRATION';
  if (successRate > 0.8 && errorRate < 0.1) return '🟠 ACCEPTABLE AI INTEGRATION';
  return '🔴 POOR AI INTEGRATION - NEEDS OPTIMIZATION';
}

function getAiRecommendations(data) {
  const recommendations = [];
  const successRate = data.metrics.successful_ai_processing?.rate || 0;
  const p95 = data.metrics.ai_processing_time?.['p(95)'] || 0;
  const avgTime = data.metrics.ai_processing_time?.avg || 0;
  const errorRate = data.metrics.ai_error_rate?.rate || 0;
  
  if (avgTime > 2000) recommendations.push('• Consider model optimization or faster inference hardware');
  if (errorRate > 0.05) recommendations.push('• Investigate AI service reliability and error handling');
  if (p95 > 4000) recommendations.push('• Implement request timeout and fallback mechanisms');
  if (successRate < 0.9) recommendations.push('• Add AI service health checks and circuit breakers');
  
  if (avgTime > 3000) recommendations.push('• Consider implementing async processing for complex AI tasks');
  if (p95 > 5000) recommendations.push('• Add AI request queuing and priority management');
  
  if (recommendations.length === 0) {
    recommendations.push('• AI integration performance is excellent!');
    recommendations.push('• Consider adding more sophisticated AI features');
  }
  
  return recommendations.join('\n');
}
