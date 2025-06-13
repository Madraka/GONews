#!/bin/bash

# 🧪 Cloudflare Tunnel Test Script
# Tests Cloudflare Tunnel connectivity and performance

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test fonksiyonu
test_endpoint() {
    local url=$1
    local description=$2
    
    echo -e "${BLUE}🧪 Testing: $description${NC}"
    echo "URL: $url"
    
    if curl -s -o /dev/null -w "%{http_code}" --max-time 10 "$url" | grep -q "200\|301\|302"; then
        echo -e "${GREEN}✅ SUCCESS${NC}"
        return 0
    else
        echo -e "${RED}❌ FAILED${NC}"
        return 1
    fi
}

# Başlık
echo -e "${BLUE}"
echo "🧪 Cloudflare Tunnel Test Suite"
echo "================================"
echo -e "${NC}"

# K8s status kontrol
echo -e "${BLUE}📊 Kubernetes Status${NC}"
echo "=========================="
kubectl get pods -n production -l app=cloudflare-tunnel
echo ""

# Tunnel status
echo -e "${BLUE}🌐 Tunnel Status${NC}"
echo "=================="
kubectl logs -n production -l app=cloudflare-tunnel --tail=5
echo ""

# Test endpoints
echo -e "${BLUE}🌍 Public Endpoint Tests${NC}"
echo "=========================="

ENDPOINTS=(
    "https://api.news.madraka.dev/health,API Health Check"
    "https://news-api.madraka.dev/health,Alternative API Health"
    "https://monitoring.news.madraka.dev,Monitoring Dashboard"
    "https://api.news.madraka.dev/api/articles,Articles Endpoint"
    "https://api.news.madraka.dev/swagger/index.html,Swagger Documentation"
)

SUCCESS_COUNT=0
TOTAL_COUNT=${#ENDPOINTS[@]}

for endpoint in "${ENDPOINTS[@]}"; do
    IFS=',' read -ra PARTS <<< "$endpoint"
    url=${PARTS[0]}
    description=${PARTS[1]}
    
    if test_endpoint "$url" "$description"; then
        ((SUCCESS_COUNT++))
    fi
    echo ""
done

# Sonuç
echo -e "${BLUE}📊 Test Results${NC}"
echo "================"
echo "Success: $SUCCESS_COUNT/$TOTAL_COUNT"

if [ $SUCCESS_COUNT -eq $TOTAL_COUNT ]; then
    echo -e "${GREEN}🎉 All tests passed! Tunnel is working perfectly!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some tests failed. Check DNS propagation and tunnel status.${NC}"
    exit 1
fi
