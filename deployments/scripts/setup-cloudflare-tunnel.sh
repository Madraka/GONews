#!/bin/bash

# 🌐 Cloudflare Tunnel K8s Quick Setup Script
# This script automatically sets up Cloudflare Tunnel in Kubernetes

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Header
echo -e "${BLUE}"
echo "🌐 Cloudflare Tunnel K8s Setup"
echo "==============================="
echo -e "${NC}"

# Pre-checks
log_info "Running pre-checks..."

# kubectl check
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl is not installed! Please install kubectl."
    exit 1
fi

# cloudflared check
if ! command -v cloudflared &> /dev/null; then
    log_warning "cloudflared is not installed. Installing..."
    if command -v brew &> /dev/null; then
        brew install cloudflared
        log_success "cloudflared installed via Homebrew"
    else
        log_error "Homebrew is not installed. Please install cloudflared manually."
        echo "curl -L --output cloudflared.pkg https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-darwin-amd64.pkg"
        echo "sudo installer -pkg cloudflared.pkg -target /"
        exit 1
    fi
fi

# Namespace check
if ! kubectl get namespace production &> /dev/null; then
    log_warning "Production namespace not found. Creating..."
    kubectl create namespace production
    log_success "Production namespace created"
fi

# Create tunnel
log_info "Creating Cloudflare tunnel..."

# Login check
if [ ! -f ~/.cloudflared/cert.pem ]; then
    log_warning "You need to login to Cloudflare..."
    cloudflared tunnel login
fi

# Check if tunnel exists
TUNNEL_NAME="news-api"
if cloudflared tunnel list | grep -q "$TUNNEL_NAME"; then
    log_warning "Tunnel '$TUNNEL_NAME' already exists"
else
    log_info "Creating new tunnel: $TUNNEL_NAME"
    cloudflared tunnel create $TUNNEL_NAME
    log_success "Tunnel created: $TUNNEL_NAME"
fi

# Get token
log_info "Getting tunnel token..."
TUNNEL_TOKEN=$(cloudflared tunnel token $TUNNEL_NAME)

if [ -z "$TUNNEL_TOKEN" ]; then
    log_error "Could not get tunnel token!"
    exit 1
fi

# Encode token to base64
TOKEN_BASE64=$(echo -n "$TUNNEL_TOKEN" | base64)

log_success "Token retrieved and encoded"

# Create or update secret
log_info "Creating/updating Kubernetes secret..."

kubectl create secret generic cloudflare-tunnel-token \
    --namespace=production \
    --from-literal=token="$TUNNEL_TOKEN" \
    --dry-run=client -o yaml | kubectl apply -f -

log_success "Secret created/updated"

# Apply deployment
log_info "Applying Cloudflare tunnel deployment..."

kubectl apply -f /Users/madraka/News/deployments/k8s/production/07-cloudflare-tunnel.yml

log_success "Deployment applied"

# Wait for pods to start
log_info "Waiting for pods to start..."
kubectl wait --for=condition=ready pod -l app=cloudflare-tunnel -n production --timeout=120s

# Get tunnel ID
TUNNEL_ID=$(cloudflared tunnel list | grep "$TUNNEL_NAME" | awk '{print $1}')

# Configuration instructions
echo -e "\n${YELLOW}🌐 Cloudflare Dashboard Configuration${NC}"
echo "==============================="
echo "Configure your tunnel at: https://one.dash.cloudflare.com/networks/tunnels"
echo ""
echo "1. Navigate to your tunnel: $TUNNEL_NAME"
echo "2. Go to 'Public Hostnames' tab"
echo "3. Add the following ingress rules:"
echo ""
echo "   Rule 1:"
echo "   ├─ Subdomain: api.news"
echo "   ├─ Domain: madraka.dev"
echo "   ├─ Service Type: HTTP"
echo "   ├─ URL: news-api-service:8080"
echo "   └─ Additional headers: Host = api.news.production"
echo ""
echo "   Rule 2:"
echo "   ├─ Subdomain: news-api"
echo "   ├─ Domain: madraka.dev"
echo "   ├─ Service Type: HTTP"
echo "   ├─ URL: news-api-service:8080"
echo "   └─ Additional headers: Host = news-api.production"
echo ""
echo "   Rule 3:"
echo "   ├─ Subdomain: monitoring.news"
echo "   ├─ Domain: madraka.dev"
echo "   ├─ Service Type: HTTP"
echo "   └─ URL: prometheus-service:9090"
echo ""
echo "4. Save the configuration"

# DNS instructions (still needed for CNAME records)
echo -e "\n${YELLOW}🌐 DNS Configuration${NC}"
echo "==============================="
echo "The DNS records will be automatically created when you configure"
echo "the public hostnames in the Cloudflare Dashboard above."
echo ""
echo "If you need to create them manually:"
echo "Type: CNAME"
echo "Target: ${TUNNEL_ID}.cfargotunnel.com"
echo "Proxy: ✅ Proxied"

# Status check
echo -e "\n${BLUE}📊 Deployment Status${NC}"
echo "==============================="
kubectl get pods -n production -l app=cloudflare-tunnel
echo ""
kubectl get svc -n production cloudflare-tunnel-metrics

# Test instructions
echo -e "\n${GREEN}🧪 Test Commands${NC}"
echo "==============================="
echo "After DNS propagation (5-10 minutes):"
echo ""
echo "curl https://api.news.madraka.dev/health"
echo "curl https://news-api.madraka.dev/health"
echo "curl https://monitoring.news.madraka.dev"

# Logs
echo -e "\n${BLUE}📋 Logs Check${NC}"
echo "==============================="
echo "kubectl logs -n production -l app=cloudflare-tunnel"

log_success "Cloudflare Tunnel setup completed!"
log_warning "Don't forget to add DNS records from Cloudflare Dashboard!"

echo -e "\n${BLUE}🔗 Useful Links${NC}"
echo "==============================="
echo "• Cloudflare Dashboard: https://dash.cloudflare.com"
echo "• Tunnel Management: https://one.dash.cloudflare.com/networks/tunnels"
echo "• DNS Management: https://dash.cloudflare.com/dns"
echo ""
