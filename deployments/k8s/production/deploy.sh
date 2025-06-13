#!/bin/bash

# Production Kubernetes Deployment Script for News API
# This script deploys the complete News API infrastructure to production

set -euo pipefail

# Configuration
NAMESPACE="production"
CONTEXT="docker-desktop"  # Using Docker Desktop Kubernetes
KUBECTL_VERSION_MIN="1.24"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed"
        exit 1
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    # Check if context exists
    if ! kubectl config get-contexts | grep -q "$CONTEXT"; then
        log_warning "Context '$CONTEXT' not found, using current context"
        CONTEXT=$(kubectl config current-context)
    fi
    
    # Check cluster connection
    if ! kubectl cluster-info --context="$CONTEXT" &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Verify Docker images exist
verify_images() {
    log_info "Verifying Docker images..."
    
    local images=("news/api:production" "news/worker:production")
    local missing_images=()
    
    for image in "${images[@]}"; do
        if ! docker image inspect "$image" &> /dev/null; then
            missing_images+=("$image")
        fi
    done
    
    if [ ${#missing_images[@]} -gt 0 ]; then
        log_error "Missing Docker images: ${missing_images[*]}"
        log_info "Building missing images..."
        
        # Build API if missing
        if [[ " ${missing_images[*]} " =~ " news/api:production " ]]; then
            log_info "Building news/api:production..."
            docker build -f deployments/dockerfiles/Dockerfile.prod -t news/api:production .
        fi
        
        # Build Worker if missing
        if [[ " ${missing_images[*]} " =~ " news/worker:production " ]]; then
            log_info "Building news/worker:production..."
            docker build -f deployments/dockerfiles/Dockerfile.worker -t news/worker:production .
        fi
    fi
    
    log_success "Docker images verified"
}

# Create namespace if it doesn't exist
create_namespace() {
    log_info "Creating namespace '$NAMESPACE'..."
    
    if kubectl get namespace "$NAMESPACE" --context="$CONTEXT" &> /dev/null; then
        log_warning "Namespace '$NAMESPACE' already exists"
    else
        kubectl create namespace "$NAMESPACE" --context="$CONTEXT"
        kubectl label namespace "$NAMESPACE" name="$NAMESPACE" --context="$CONTEXT"
        log_success "Namespace '$NAMESPACE' created"
    fi
}

# Deploy resources in order
deploy_resources() {
    log_info "Deploying resources to production..."
    
    local deploy_order=(
        "08-resources.yml"
        "01-secrets.yml" 
        "02-configmaps.yml"
        "03-databases.yml"
        "04-api.yml"
        "05-worker.yml"
        "06-ingress.yml"
        "07-monitoring.yml"
    )
    
    for file in "${deploy_order[@]}"; do
        local filepath="deployments/k8s/production/$file"
        
        if [ -f "$filepath" ]; then
            log_info "Deploying $file..."
            kubectl apply -f "$filepath" --context="$CONTEXT"
            log_success "Deployed $file"
        else
            log_warning "File $filepath not found, skipping"
        fi
    done
}

# Wait for deployments to be ready
wait_for_deployments() {
    log_info "Waiting for deployments to be ready..."
    
    local deployments=("postgres" "redis" "news-api" "news-worker")
    
    for deployment in "${deployments[@]}"; do
        log_info "Waiting for $deployment to be ready..."
        kubectl rollout status deployment/"$deployment" -n "$NAMESPACE" --context="$CONTEXT" --timeout=300s
        log_success "$deployment is ready"
    done
}

# Verify deployment health
verify_deployment() {
    log_info "Verifying deployment health..."
    
    # Check pod status
    log_info "Pod status:"
    kubectl get pods -n "$NAMESPACE" --context="$CONTEXT" -o wide
    
    # Check service status
    log_info "Service status:"
    kubectl get services -n "$NAMESPACE" --context="$CONTEXT"
    
    # Check ingress status
    log_info "Ingress status:"
    kubectl get ingress -n "$NAMESPACE" --context="$CONTEXT"
    
    # Check HPA status
    log_info "HPA status:"
    kubectl get hpa -n "$NAMESPACE" --context="$CONTEXT"
    
    # Health check
    log_info "Running health checks..."
    
    # Wait for API to be responsive
    local api_pod=$(kubectl get pods -n "$NAMESPACE" -l app=news-api --context="$CONTEXT" -o jsonpath='{.items[0].metadata.name}')
    if [ -n "$api_pod" ]; then
        if kubectl exec -n "$NAMESPACE" "$api_pod" --context="$CONTEXT" -- wget -q --spider http://localhost:8081/health; then
            log_success "API health check passed"
        else
            log_warning "API health check failed"
        fi
    fi
    
    log_success "Deployment verification completed"
}

# Display connection information
show_connection_info() {
    log_info "Deployment completed successfully!"
    echo
    echo "=== Connection Information ==="
    echo
    
    # Get LoadBalancer IP if available
    local lb_ip=$(kubectl get service -n ingress-nginx ingress-nginx-controller --context="$CONTEXT" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    
    if [ -n "$lb_ip" ]; then
        echo "API Endpoint: https://$lb_ip"
        echo "Add the following to your /etc/hosts file for domain access:"
        echo "$lb_ip api.news.production"
        echo "$lb_ip news-api.production"
    else
        echo "LoadBalancer IP not available. Use port-forward for testing:"
        echo "kubectl port-forward -n $NAMESPACE service/news-api-service 8081:8081 --context=$CONTEXT"
    fi
    
    echo
    echo "=== Monitoring ==="
    echo "Metrics endpoint: http://localhost:9090/metrics (via port-forward)"
    echo "kubectl port-forward -n $NAMESPACE service/news-api-service 9090:9090 --context=$CONTEXT"
    echo
    echo "=== Management Commands ==="
    echo "View logs: kubectl logs -n $NAMESPACE -l app=news-api --context=$CONTEXT"
    echo "Scale API: kubectl scale deployment news-api --replicas=10 -n $NAMESPACE --context=$CONTEXT"
    echo "Scale Worker: kubectl scale deployment news-worker --replicas=8 -n $NAMESPACE --context=$CONTEXT"
    echo
}

# Rollback function
rollback() {
    log_warning "Rolling back deployment..."
    
    local deployments=("news-api" "news-worker")
    
    for deployment in "${deployments[@]}"; do
        kubectl rollout undo deployment/"$deployment" -n "$NAMESPACE" --context="$CONTEXT"
        log_info "Rolled back $deployment"
    done
    
    log_success "Rollback completed"
}

# Cleanup function
cleanup() {
    log_warning "Cleaning up resources..."
    
    read -p "Are you sure you want to delete all resources in namespace '$NAMESPACE'? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl delete namespace "$NAMESPACE" --context="$CONTEXT"
        log_success "Cleanup completed"
    else
        log_info "Cleanup cancelled"
    fi
}

# Main deployment function
main() {
    local action="${1:-deploy}"
    
    case "$action" in
        "deploy")
            check_prerequisites
            verify_images
            create_namespace
            deploy_resources
            wait_for_deployments
            verify_deployment
            show_connection_info
            ;;
        "rollback")
            rollback
            ;;
        "cleanup")
            cleanup
            ;;
        "verify")
            verify_deployment
            ;;
        "info")
            show_connection_info
            ;;
        *)
            echo "Usage: $0 {deploy|rollback|cleanup|verify|info}"
            echo
            echo "Commands:"
            echo "  deploy   - Deploy the complete News API infrastructure"
            echo "  rollback - Rollback to previous deployment"
            echo "  cleanup  - Delete all resources (WARNING: destructive)"
            echo "  verify   - Verify current deployment status"
            echo "  info     - Show connection information"
            exit 1
            ;;
    esac
}

# Handle script interruption
trap 'log_error "Deployment interrupted"; exit 1' INT TERM

# Run main function
main "$@"
