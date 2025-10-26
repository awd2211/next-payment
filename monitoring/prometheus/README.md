# Prometheus Configuration for BFF Services

Complete Prometheus monitoring setup for Admin BFF and Merchant BFF services.

## 📋 Overview

This directory contains:
- **prometheus.yml** - Main Prometheus configuration
- **alerts/bff-alerts.yml** - 21 alert rules for BFF services
- **rules/bff-recording-rules.yml** - 25 recording rules for performance optimization

## 🚀 Quick Start

### Method 1: Docker Compose (Recommended)

```bash
# Start Prometheus with BFF services
docker-compose -f docker-compose.yml up -d prometheus

# Verify Prometheus is running
curl http://localhost:40090/-/healthy

# Access Prometheus UI
open http://localhost:40090
```

**Docker Compose Configuration**:
```yaml
prometheus:
  image: prom/prometheus:latest
  ports:
    - "40090:9090"
  volumes:
    - ./monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
    - ./monitoring/prometheus/alerts:/etc/prometheus/alerts
    - ./monitoring/prometheus/rules:/etc/prometheus/rules
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.path=/prometheus'
    - '--web.console.libraries=/usr/share/prometheus/console_libraries'
    - '--web.console.templates=/usr/share/prometheus/consoles'
    - '--web.enable-lifecycle'
  restart: unless-stopped
```

### Method 2: Manual Installation

```bash
# 1. Download Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar -xzf prometheus-2.45.0.linux-amd64.tar.gz
cd prometheus-2.45.0.linux-amd64

# 2. Copy configuration
cp /home/eric/payment/monitoring/prometheus/prometheus.yml ./
cp -r /home/eric/payment/monitoring/prometheus/alerts ./
cp -r /home/eric/payment/monitoring/prometheus/rules ./

# 3. Start Prometheus
./prometheus --config.file=./prometheus.yml --web.listen-address=:40090

# 4. Verify
curl http://localhost:40090/api/v1/status/config
```

## 📊 Monitoring Targets

### BFF Services

| Service | Endpoint | Scrape Interval | Labels |
|---------|----------|-----------------|--------|
| Admin BFF | http://localhost:40001/metrics | 10s | `service=admin-bff-service`, `security_level=high` |
| Merchant BFF | http://localhost:40023/metrics | 15s | `service=merchant-bff-service`, `security_level=medium` |

**Admin BFF** has shorter scrape interval (10s) for enhanced security monitoring.

### Health Check

```bash
# Check if targets are up
curl http://localhost:40090/api/v1/targets | jq '.data.activeTargets[] | {job, health, lastError}'

# Expected output:
# {
#   "job": "admin-bff",
#   "health": "up",
#   "lastError": ""
# }
```

## 🔔 Alert Rules (21 Total)

Located in `alerts/bff-alerts.yml`:

### Critical Alerts (Severity: critical)

1. **BFFServiceDown** - Service unavailable for 1 minute
2. **BFFHighErrorRate** - Error rate >5% for 5 minutes
3. **BFFExtremelyHighLatency** - P95 latency >3s for 5 minutes
4. **BFFMemoryExhaustion** - Memory usage >90% for 5 minutes
5. **BFFHighRateLimitViolations** - Rate limit abuse (>10 req/s for 5 min)
6. **BFFCriticalSecurityEvents** - Security events >50/min for 5 minutes

### Warning Alerts (Severity: warning)

7. **BFFHighLatency** - P95 latency >1s for 5 minutes
8. **BFFHighMemoryUsage** - Memory usage >80% for 10 minutes
9. **BFFHighCPUUsage** - CPU usage >80% for 10 minutes
10. **BFFMediumRateLimitViolations** - Moderate rate limit violations
11. **BFFAuthFailures** - High 401 errors (>5/min for 5 min)
12. **BFFPermissionDenied** - High 403 errors (>5/min for 5 min)
13. **BFF2FAFailures** - Admin BFF 2FA failures (>10/min for 5 min)
14. **BFFHighGoroutines** - Goroutine leak (>1000 for 10 min)
15. **BFFSlowRequests** - P99 latency >5s for 5 minutes
16. **BFFHighRequestSize** - Average request >1MB for 10 minutes
17. **BFFHighResponseSize** - Average response >5MB for 10 minutes

### Info Alerts (Severity: info)

18. **BFFServiceRestarted** - Service restarted in last 5 minutes
19. **BFFLowTraffic** - Request rate <1 req/min for 30 minutes
20. **BFFUnusualErrorPattern** - 4xx errors >20% for 10 minutes
21. **BFFFileDescriptorWarning** - FD usage >70% for 10 minutes

**View Active Alerts**:
```bash
curl http://localhost:40090/api/v1/alerts | jq '.data.alerts[] | {alertname, state, value}'
```

## 📈 Recording Rules (25 Total)

Located in `rules/bff-recording-rules.yml`:

Recording rules pre-calculate frequently used queries to improve dashboard performance.

### HTTP Performance Metrics (7 rules)

```promql
# 1. Request rate (req/s) by service
job:http_requests:rate5m

# 2. Request rate by status code
job:http_requests:rate5m:by_status

# 3. Request rate by endpoint
job:http_requests:rate5m:by_path

# 4. Error rate (5xx)
job:http_errors:rate5m

# 5. P95 latency
job:http_request_duration:p95

# 6. P99 latency
job:http_request_duration:p99

# 7. Average latency
job:http_request_duration:avg
```

### Security Metrics (5 rules)

```promql
# 8. Rate limit violations (429)
job:rate_limit_violations:rate5m

# 9. Authentication failures (401)
job:auth_failures:rate5m

# 10. Permission denied (403)
job:permission_denied:rate5m

# 11. 2FA failures (Admin BFF only)
job:twofa_failures:rate5m

# 12. Overall security events (401 + 403 + 429)
job:security_events:rate5m
```

### Resource Usage Metrics (4 rules)

```promql
# 13. Memory usage
job:memory_usage:bytes

# 14. CPU usage rate
job:cpu_usage:rate5m

# 15. Goroutine count
job:goroutines:current

# 16. File descriptor usage ratio
job:fd_usage:ratio
```

### Business Metrics (2 rules)

```promql
# 19. Average request size
job:http_request_size:avg

# 20. Average response size
job:http_response_size:avg
```

### SLI/SLO Metrics (3 rules)

```promql
# 21. Availability (success rate %)
job:availability:rate5m

# 22. Latency SLI (P95 < 500ms)
job:latency_sli:p95_lt_500ms

# 23. Error budget remaining (%)
job:error_budget:remaining
```

### Health Metrics (2 rules)

```promql
# 24. Service status (1 = UP, 0 = DOWN)
job:up:status

# 25. Restart count (last 1 hour)
job:restarts:increase1h
```

## 🔍 Useful Queries

### Service Health

```promql
# Service availability
up{job=~"admin-bff|merchant-bff"}

# Service uptime
time() - process_start_time_seconds{job="admin-bff"}

# Request success rate
job:availability:rate5m{job="admin-bff"}
```

### Performance

```promql
# P95 latency by service
job:http_request_duration:p95

# Top 10 slowest endpoints
topk(10,
  histogram_quantile(0.95,
    rate(http_request_duration_seconds_bucket{job="admin-bff"}[5m])
  ) by (path)
)

# Request rate by endpoint
job:http_requests:rate5m:by_path{job="admin-bff"}

# Error rate trend
job:http_errors:rate5m
```

### Security

```promql
# Rate limit violations per minute
rate(http_requests_total{status="429"}[1m]) * 60

# Authentication failures by IP (requires client_ip label)
sum(rate(http_requests_total{status="401"}[5m])) by (client_ip)

# 2FA failure rate (Admin BFF)
job:twofa_failures:rate5m

# Security events timeline
job:security_events:rate5m
```

### Resource Usage

```promql
# Memory usage percentage (assuming limit is 512MB for Admin BFF)
(job:memory_usage:bytes{job="admin-bff"} / (512 * 1024 * 1024)) * 100

# CPU cores used
job:cpu_usage:rate5m

# Goroutine growth rate
deriv(job:goroutines:current{job="admin-bff"}[5m])

# File descriptor usage
job:fd_usage:ratio * 100
```

### Business Analytics

```promql
# Total requests per hour
sum(increase(http_requests_total{job="admin-bff"}[1h]))

# Request distribution by status code
sum(rate(http_requests_total{job="admin-bff"}[5m])) by (status)

# Average request/response size ratio
job:http_response_size:avg / job:http_request_size:avg

# Traffic volume (bytes/s)
rate(http_request_size_bytes_sum{job="admin-bff"}[5m]) +
rate(http_response_size_bytes_sum{job="admin-bff"}[5m])
```

## 🧪 Testing

### Validate Configuration

```bash
# Check syntax
promtool check config monitoring/prometheus/prometheus.yml

# Check alert rules
promtool check rules monitoring/prometheus/alerts/bff-alerts.yml

# Check recording rules
promtool check rules monitoring/prometheus/rules/bff-recording-rules.yml
```

### Test Query Performance

```bash
# Query API
curl -G http://localhost:40090/api/v1/query \
  --data-urlencode 'query=job:http_requests:rate5m' | jq .

# Query range
curl -G http://localhost:40090/api/v1/query_range \
  --data-urlencode 'query=job:http_request_duration:p95' \
  --data-urlencode 'start=2025-10-26T00:00:00Z' \
  --data-urlencode 'end=2025-10-26T12:00:00Z' \
  --data-urlencode 'step=60s' | jq .
```

### Trigger Test Alerts

```bash
# Stop Admin BFF to trigger BFFServiceDown alert
pkill -f admin-bff-service

# Wait 1-2 minutes
sleep 120

# Check firing alerts
curl http://localhost:40090/api/v1/alerts | jq '.data.alerts[] | select(.state=="firing")'

# Restart service
cd /home/eric/payment/backend && ./scripts/start-bff-services.sh
```

## 🛠️ Troubleshooting

### Issue 1: Targets are Down

**Symptom**: `up{job="admin-bff"} = 0`

**Diagnosis**:
```bash
# Check target health
curl http://localhost:40090/api/v1/targets | jq '.data.activeTargets[] | select(.health=="down")'

# Check BFF service
curl http://localhost:40001/metrics
curl http://localhost:40023/metrics

# Check BFF logs
tail -f /home/eric/payment/backend/logs/bff/admin-bff.log
```

**Solution**:
- Verify BFF services are running: `ps aux | grep bff-service`
- Check port bindings: `netstat -tulpn | grep -E "40001|40023"`
- Restart services: `./backend/scripts/start-bff-services.sh`

### Issue 2: Recording Rules Not Evaluating

**Symptom**: Recording rule metrics missing

**Diagnosis**:
```bash
# Check rule evaluation
curl http://localhost:40090/api/v1/rules | jq '.data.groups[] | select(.name=="bff_http_performance")'

# Check for errors
docker logs payment-prometheus 2>&1 | grep -i error
```

**Solution**:
- Validate rule syntax: `promtool check rules monitoring/prometheus/rules/bff-recording-rules.yml`
- Reload Prometheus: `curl -X POST http://localhost:40090/-/reload`
- Check evaluation interval in prometheus.yml

### Issue 3: Alerts Not Firing

**Symptom**: Expected alerts not appearing

**Diagnosis**:
```bash
# Check alert state
curl http://localhost:40090/api/v1/alerts | jq '.data.alerts[] | {alertname, state, activeAt, value}'

# Check rule evaluation
curl http://localhost:40090/api/v1/rules | jq '.data.groups[] | select(.file=="alerts/bff-alerts.yml")'
```

**Solution**:
- Verify alert thresholds match actual values
- Check `for` duration hasn't expired
- Ensure Alertmanager is configured (if using)
- Reload alerts: `curl -X POST http://localhost:40090/-/reload`

### Issue 4: High Cardinality

**Symptom**: Prometheus using excessive memory

**Diagnosis**:
```bash
# Check metric cardinality
curl http://localhost:40090/api/v1/status/tsdb | jq '.data.numSeries'

# Check label cardinality
curl http://localhost:40090/api/v1/label/__name__/values | jq '.data | length'
```

**Solution**:
- Avoid high-cardinality labels (user_id, trace_id, etc.)
- Use metric relabeling to drop unnecessary labels
- Increase Prometheus memory: `--storage.tsdb.retention.size=10GB`

## 📚 Integration with Other Tools

### Grafana

Import dashboards using this Prometheus datasource:

```yaml
# Grafana datasource configuration
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://localhost:40090
    isDefault: true
```

See [/home/eric/payment/monitoring/grafana/README.md](../grafana/README.md) for dashboard setup.

### Alertmanager (Optional)

Configure Alertmanager for alert routing:

```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@payment-platform.com'
  smtp_auth_username: 'alerts@payment-platform.com'
  smtp_auth_password: 'your-password'

route:
  group_by: ['alertname', 'job']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: 'team-ops'

  routes:
    - match:
        severity: critical
      receiver: 'pagerduty'

    - match:
        severity: warning
      receiver: 'slack'

receivers:
  - name: 'team-ops'
    email_configs:
      - to: 'ops-team@payment-platform.com'

  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
        channel: '#alerts'

  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_KEY'
```

### Jaeger (Tracing Correlation)

Link metrics with traces using trace_id:

```promql
# Query requests with high latency and their trace IDs (requires trace_id label)
topk(10,
  http_request_duration_seconds{job="admin-bff"} > 1
) by (trace_id)
```

## 🔗 Related Documentation

- **[BFF Services README](../../backend/services/README.BFF.md)** - Quick start guide
- **[BFF Security Summary](../../BFF_SECURITY_COMPLETE_SUMMARY.md)** - Architecture overview
- **[Grafana Dashboard Guide](../grafana/README.md)** - Visualization setup
- **[Prometheus Official Docs](https://prometheus.io/docs/prometheus/latest/getting_started/)** - Official documentation

## 📝 Best Practices

### Production Configuration

1. **Increase retention**:
```yaml
# In prometheus.yml or command-line flags
--storage.tsdb.retention.time=30d
--storage.tsdb.retention.size=50GB
```

2. **Enable remote write** (for long-term storage):
```yaml
remote_write:
  - url: "http://victoria-metrics:8428/api/v1/write"
    queue_config:
      max_samples_per_send: 10000
```

3. **Tune scrape intervals**:
- Critical services: 10-15s
- Normal services: 30s
- Exporters: 60s

4. **Use recording rules** for expensive queries in dashboards

5. **Set up Alertmanager** for proper alert routing and deduplication

### Security

1. **Enable authentication**:
```yaml
# prometheus.yml
basic_auth_users:
  admin: $2y$10$...hashed_password...
```

2. **Use HTTPS** with TLS certificates

3. **Restrict access** using firewall rules:
```bash
# Only allow localhost and monitoring subnet
iptables -A INPUT -p tcp --dport 40090 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 40090 -s 10.0.0.0/8 -j ACCEPT
iptables -A INPUT -p tcp --dport 40090 -j DROP
```

4. **Disable public endpoints** in production:
```yaml
--web.enable-admin-api=false
--web.enable-lifecycle=false
```

---

**Last Updated**: 2025-10-26
**Maintainer**: Payment Platform Team
**Support**: https://github.com/your-org/payment-platform/issues
