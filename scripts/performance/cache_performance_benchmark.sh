#!/bin/bash

# Cache Performance Benchmark Script
# Tests optimized cache performance under different load scenarios

echo "🚀 Cache Performance Benchmark Started"
echo "======================================"

BASE_URL="http://localhost:8081"
RESULTS_FILE="cache_benchmark_results_$(date +%Y%m%d_%H%M%S).json"

# Function to get cache analytics
get_cache_stats() {
    curl -s "$BASE_URL/api/cache/analytics" | jq '.analytics.performance_metrics'
}

# Function to make concurrent requests
make_concurrent_requests() {
    local num_requests=$1
    local concurrency=$2
    echo "📊 Making $num_requests requests with $concurrency concurrency..."
    
    # Use GNU parallel if available, otherwise use background processes
    if command -v parallel >/dev/null 2>&1; then
        seq 1 $num_requests | parallel -j$concurrency "curl -s $BASE_URL/api/articles > /dev/null"
    else
        for ((i=1; i<=num_requests; i++)); do
            curl -s "$BASE_URL/api/articles" > /dev/null &
            if (( i % concurrency == 0 )); then
                wait
            fi
        done
        wait
    fi
}

# Test scenarios
echo "🧪 Test 1: Cold Cache Performance"
echo "================================"
echo "Initial cache stats:"
get_cache_stats | jq '.'

echo ""
echo "🔥 Test 2: Cache Warming (20 sequential requests)"
echo "=============================================="
for i in {1..20}; do
    curl -s "$BASE_URL/api/articles" > /dev/null
done

echo "Cache stats after warming:"
get_cache_stats | jq '.'

echo ""
echo "⚡ Test 3: High Concurrency Load (50 requests, 10 concurrent)"
echo "========================================================="
make_concurrent_requests 50 10

echo "Cache stats after high concurrency:"
get_cache_stats | jq '.'

echo ""
echo "🚀 Test 4: Stress Test (100 requests, 20 concurrent)"
echo "================================================"
make_concurrent_requests 100 20

echo "Final cache stats:"
get_cache_stats | jq '.' | tee "$RESULTS_FILE"

echo ""
echo "📈 Performance Summary"
echo "====================="
final_stats=$(get_cache_stats)

overall_hit_rate=$(echo "$final_stats" | jq -r '.overall_hit_rate')
l1_hit_ratio=$(echo "$final_stats" | jq -r '.l1_hit_ratio')
l2_hit_ratio=$(echo "$final_stats" | jq -r '.l2_hit_ratio')
avg_latency_l1=$(echo "$final_stats" | jq -r '.avg_latency_l1')
avg_latency_l2=$(echo "$final_stats" | jq -r '.avg_latency_l2')
efficiency=$(echo "$final_stats" | jq -r '.overall_efficiency')

echo "✅ Overall Hit Rate: $(echo "$overall_hit_rate * 100" | bc -l | cut -d. -f1)%"
echo "✅ L1 Hit Ratio: $(echo "$l1_hit_ratio * 100" | bc -l | cut -d. -f1)%"
echo "✅ L2 Hit Ratio: $(echo "$l2_hit_ratio * 100" | bc -l | cut -d. -f1)%"
echo "⚡ L1 Average Latency: $avg_latency_l1"
echo "⚡ L2 Average Latency: $avg_latency_l2"
echo "🏆 Overall Efficiency: $efficiency"

echo ""
echo "📁 Results saved to: $RESULTS_FILE"
echo "🎯 Benchmark completed successfully!"

# Cache health check
echo ""
echo "🩺 Cache Health Check"
echo "===================="
curl -s "$BASE_URL/api/cache/health" | jq '.overall_healthy, .status'
