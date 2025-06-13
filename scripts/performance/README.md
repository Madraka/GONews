# Performance Testing & Optimization Suite

This directory contains comprehensive performance testing tools and scripts for the GONews API project, implementing **Phase 1** of the [Performance Optimization Plan](../docs/performance_optimization_plan.md).

## 📋 Overview

The performance testing suite provides:
- **Comprehensive load testing** scenarios for different traffic patterns
- **Database performance analysis** and optimization recommendations
- **Performance baseline establishment** for comparison
- **Automated reporting** and metrics collection
- **Monitoring integration** with Prometheus and Grafana

## 🛠️ Prerequisites

### Required Tools
- **k6** - Load testing tool ([Installation Guide](https://k6.io/docs/getting-started/installation/))
- **PostgreSQL Client** - For database analysis (`psql`)
- **curl** - For API testing
- **jq** - For JSON processing (optional, improves reporting)

### Installation Commands

```bash
# macOS
brew install k6 postgresql jq

# Ubuntu/Debian
sudo apt-get install postgresql-client jq
wget https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.deb
sudo dpkg -i k6-v0.47.0-linux-amd64.deb

# Windows (using Chocolatey)
choco install k6 postgresql jq
```

## 📁 Directory Structure

```
scripts/
├── performance/                    # Load testing scenarios
│   ├── load-test-normal.js        # Normal traffic patterns
│   ├── load-test-spike.js         # Traffic spike scenarios
│   ├── load-test-database.js      # Database-focused testing
│   └── load-test-ai.js            # AI integration testing
├── run-performance-tests.sh       # Main test runner
├── establish-baseline.sh          # Performance baseline tool
├── analyze-database.sh            # Database analysis tool
└── README.md                      # This file

reports/
├── performance/                    # Load test results
├── baseline/                      # Baseline measurements
├── database/                      # Database analysis reports
└── monitoring/                    # Monitoring data
```

## 🚀 Quick Start

### 1. Start the GONews API
```bash
# Ensure the API is running
curl http://localhost:8080/api/news
```

### 2. Establish Performance Baseline
```bash
./establish-baseline.sh
```

### 3. Run Performance Tests
```bash
# Run all performance tests
./run-performance-tests.sh

# Or run specific test types
./run-performance-tests.sh normal     # Normal traffic only
./run-performance-tests.sh spike      # Traffic spike only
./run-performance-tests.sh database   # Database performance only
```

### 4. Analyze Database Performance
```bash
./analyze-database.sh
```

## 📊 Test Scenarios

### 1. Normal Traffic Test (`load-test-normal.js`)
**Purpose:** Simulate realistic user browsing patterns
- **Users:** 10 → 50 → 50 → 0 (over 19 minutes)
- **Patterns:** Homepage browsing, category filtering, search queries
- **Targets:** P95 < 200ms, Error rate < 1%, >100 RPS

**Usage:**
```bash
k6 run scripts/performance/load-test-normal.js
```

### 2. Traffic Spike Test (`load-test-spike.js`)
**Purpose:** Test system resilience during breaking news events
- **Users:** 50 → 300 → 500 → 300 → 50 → 0 (over 5 minutes)
- **Patterns:** Aggressive browsing, rapid requests, breaking news queries
- **Targets:** P95 < 500ms, Error rate < 5%, >200 RPS

**Usage:**
```bash
k6 run scripts/performance/load-test-spike.js
```

### 3. Database Performance Test (`load-test-database.js`)
**Purpose:** Focus on database query optimization
- **Users:** 20 → 80 → 120 → 80 → 0 (over 12 minutes)
- **Patterns:** Complex queries, deep pagination, analytics queries
- **Targets:** DB queries P95 < 300ms, Cache hit rate > 60%

**Usage:**
```bash
k6 run scripts/performance/load-test-database.js
```

### 4. AI Integration Test (`load-test-ai.js`)
**Purpose:** Test AI processing performance
- **Users:** 5 → 15 → 25 → 15 → 0 (over 9 minutes)
- **Patterns:** Text summarization, content analysis, classification
- **Targets:** AI processing P95 < 3000ms, Success rate > 90%

**Usage:**
```bash
k6 run scripts/performance/load-test-ai.js
```

## 📈 Performance Metrics

### Key Performance Indicators (KPIs)
| Metric | Target | Normal | Spike | Database | AI |
|--------|--------|--------|-------|----------|----| 
| Response Time P95 | <200ms | ✓ | <500ms | <300ms | <3000ms |
| Error Rate | <1% | ✓ | <5% | <0.1% | <5% |
| Requests/sec | >100 | ✓ | >200 | - | - |
| Cache Hit Rate | >60% | - | - | ✓ | - |
| AI Success Rate | >90% | - | - | - | ✓ |

### Custom Metrics Tracked
- **Database query time** - Specific database performance
- **Cache hit/miss rates** - Caching effectiveness
- **AI processing time** - AI endpoint performance
- **Connection errors** - Database stability
- **Timeout rates** - System stability under load

## 📋 Reports and Analysis

### Automated Reports
Each test generates comprehensive reports in JSON and text formats:

1. **Performance Summary** - High-level metrics and status
2. **Detailed Metrics** - Complete k6 output with all timing data
3. **Recommendations** - Specific optimization suggestions
4. **Baseline Comparison** - Performance changes over time

### Report Locations
```
reports/
├── performance/
│   ├── normal-traffic-summary.json
│   ├── spike-traffic-summary.json
│   ├── database-performance-summary.json
│   └── performance_summary_YYYYMMDD_HHMMSS.md
├── baseline/
│   ├── system_baseline_YYYYMMDD_HHMMSS.txt
│   ├── api_baseline_YYYYMMDD_HHMMSS.txt
│   └── baseline_report_YYYYMMDD_HHMMSS.md
└── database/
    ├── db_analysis_YYYYMMDD_HHMMSS.txt
    └── db_recommendations_YYYYMMDD_HHMMSS.md
```

## 🔧 Configuration

### API Keys for Testing
The load tests use different API tier keys:
- `api_key_basic_1234` - Basic tier (rate limited)
- `api_key_pro_5678` - Pro tier (higher limits)
- `api_key_enterprise_9012` - Enterprise tier (unlimited)

### Database Connection
Configure database analysis in `analyze-database.sh`:
```bash
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="gonews"
DB_USER="postgres"
DB_PASSWORD=""  # Set as environment variable if needed
```

### Test Thresholds
Modify thresholds in individual test files:
```javascript
export const options = {
  thresholds: {
    http_req_duration: ['p(95)<200'],    // 95% < 200ms
    http_req_failed: ['rate<0.01'],      // Error rate < 1%
    http_reqs: ['rate>100'],             // > 100 RPS
  },
};
```

## 🔍 Monitoring Integration

### Prometheus Metrics
Tests automatically collect and analyze:
- Request duration histograms
- Error rate counters
- Custom application metrics
- Database connection stats

### Grafana Dashboards
Compatible with existing Grafana dashboards:
- **API Performance** - Response times and error rates
- **Database Performance** - Query times and connection pools
- **System Resources** - CPU, memory, and disk utilization

### Real-time Monitoring
```bash
# View metrics during testing
curl http://localhost:8080/metrics

# Prometheus queries
http://localhost:9090

# Grafana dashboards
http://localhost:3000
```

## 🐛 Troubleshooting

### Common Issues

#### 1. k6 Not Found
```bash
# Install k6
brew install k6  # macOS
# or follow installation guide for your OS
```

#### 2. API Not Running
```bash
# Check API status
curl http://localhost:8080/api/news

# Start API if needed
cd /path/to/gonews
go run main.go
```

#### 3. Database Connection Failed
```bash
# Check PostgreSQL status
psql -h localhost -p 5432 -d gonews -U postgres -c "SELECT version();"

# Set password if needed
export DB_PASSWORD="your_password"
```

#### 4. Permission Denied on Scripts
```bash
# Make scripts executable
chmod +x *.sh
```

### Performance Issues

#### High Response Times
1. Check database query performance
2. Verify cache hit rates
3. Monitor system resources (CPU, memory)
4. Review application logs for errors

#### High Error Rates
1. Check rate limiting configuration
2. Verify API key validity
3. Monitor database connection pool
4. Check external service dependencies

#### Low Throughput
1. Optimize database indexes
2. Tune connection pool settings
3. Implement response caching
4. Scale horizontally if needed

## 📚 Advanced Usage

### Custom Test Scenarios
Create custom load tests by copying existing scenarios:
```bash
cp scripts/performance/load-test-normal.js scripts/performance/load-test-custom.js
# Edit the new file for your specific needs
k6 run scripts/performance/load-test-custom.js
```

### Continuous Integration
Integrate with CI/CD pipelines:
```yaml
# GitHub Actions example
- name: Run Performance Tests
  run: |
    ./scripts/establish-baseline.sh
    ./scripts/run-performance-tests.sh
    # Upload reports to artifacts
```

### Load Testing Best Practices
1. **Start with baseline** - Always establish baseline before testing
2. **Incremental load** - Gradually increase load to find limits
3. **Monitor resources** - Watch CPU, memory, and database during tests
4. **Test in isolation** - Run tests on dedicated test environment
5. **Regular testing** - Include performance tests in CI/CD pipeline

## 📞 Support

For issues or questions about performance testing:

1. **Check the logs** in `logs/performance/`
2. **Review the reports** in `reports/`
3. **Consult the performance plan** in `docs/performance_optimization_plan.md`
4. **Open an issue** with test results and error logs

## 🎯 Next Steps

After running performance tests:

1. **Analyze Results** - Review all generated reports
2. **Implement Optimizations** - Follow database and application recommendations
3. **Re-test** - Validate improvements with new performance tests
4. **Monitor Continuously** - Set up automated performance monitoring
5. **Scale Planning** - Plan horizontal scaling based on load limits

---

**Performance Testing Suite v1.0**  
*Part of GONews API Performance Optimization Plan*  
*Generated: 2025-05-28*
