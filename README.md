# Go Load Balancer

A production-ready HTTP load balancer implementation in Go with health checking, metrics collection, and pluggable load balancing strategies.

## Features

- **Round-robin load balancing** with atomic thread-safe operations
- **Health checking** with automatic failure detection and recovery
- **Prometheus metrics** endpoint for observability
- **Strategy pattern** for pluggable load balancing algorithms
- **Configuration validation** with comprehensive error checking
- **Context-aware timeouts** for backend requests and health checks
- **Graceful error handling** with structured logging

## Quick Start

```bash
# Run all tests
make test

# Start backends manually
make run-backends

# Start load balancer manually  
make run-lb

# Test requests
curl http://localhost:8000/          # Load balanced requests
curl http://localhost:8000/metrics   # Prometheus metrics
```

## Architecture

```
internal/
├── balancer/     # Core load balancing logic with strategy pattern
├── config/       # Configuration management and validation
├── pool/         # Backend server pool with health tracking
├── healthcheck/  # Periodic health monitoring system
├── strategy/     # Load balancing algorithms (round-robin, etc.)
└── metrics/      # Prometheus metrics collection
```

## Configuration

```bash
go run main.go \
  -port=8000 \
  -backends="http://backend1:8080,http://backend2:8081" \
  -health-interval=10 \
  -health-timeout=2 \
  -backend-timeout=30
```

## Metrics

The load balancer exposes Prometheus-compatible metrics at `/metrics`:

```
go_balancer_requests_total 42
go_balancer_requests_success_total 40
go_balancer_backend_requests_total{backend="backend-1"} 14
go_balancer_backend_healthy{state="healthy"} 3
```

## Key Design Patterns

- **Strategy Pattern**: Pluggable load balancing algorithms
- **Interface Segregation**: `LoadBalancingStrategy`, `MetricsProvider` interfaces
- **Configuration Builder**: Validation with comprehensive error reporting
- **Thread Safety**: `sync.RWMutex` for backend pool, atomic counters for round-robin
- **Context Propagation**: Request timeouts and cancellation support

## Available Commands

```bash
make help           # Show all available commands
make test           # Run automated test suite
make run-backends   # Start test backend servers
make run-lb         # Start load balancer
make kill-processes # Clean up all running processes
make check-ports    # Check port usage status
```

## Production Readiness

This implementation includes enterprise-grade features:

- **Health checking** with configurable intervals and timeouts
- **Metrics collection** ready for Prometheus scraping
- **Input validation** preventing invalid configurations
- **Graceful degradation** when backends fail
- **Comprehensive logging** for debugging and monitoring
- **Clean shutdown** handling for production deployments

Built following production Go patterns with proper error handling, concurrent safety, and observability.

## Future Improvements

### Production Environment & Containerization
**Goal**: Create a complete production-like environment using Docker

**Planned Infrastructure**:
- **Multi-stage Dockerfile** with optimized production builds and security scanning
- **Docker Compose** orchestrating the complete stack with service discovery
- **Prometheus** for metrics collection and alerting rules
- **Thanos** for long-term metrics storage and global query capabilities  
- **Grafana** with custom dashboards for load balancer and infrastructure metrics
- **Mock client services** generating realistic traffic patterns and load testing
- **Service mesh integration** (Envoy/Istio) for advanced traffic management
- **Distributed tracing** with Jaeger for request flow visualization

**Advanced Features**:
- **Circuit breaker pattern** with configurable failure thresholds
- **Rate limiting** with token bucket and sliding window algorithms
- **Weighted load balancing** with dynamic backend weighting
- **Health check diversity** (HTTP, TCP, gRPC, custom scripts)
- **Configuration hot-reloading** without service interruption
- **Admin API** for runtime backend management and configuration updates