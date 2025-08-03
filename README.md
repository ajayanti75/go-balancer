# Go Load Balancer

A production-ready HTTP load balancer implementation in Go with health checking, metrics collection, and pluggable load balancing strategies.

## Features

- **Round-robin load balancing** with atomic thread-safe operations
- **Health checking** with automatic failure detection and recovery
- **Prometheus metrics** endpoint for observability
- **Strategy pattern** for pluggable load balancing algorithms
- **Configuration validation** with comprehensive error checking
- **Context-aware timeouts** for backend requests and health checks
- **Graceful error handling** with structured error types and comprehensive context

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
├── metrics/      # Prometheus metrics collection
└── errors/       # Structured error types with context and HTTP mapping
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

## Error Handling

The load balancer uses structured error types with specific error codes and HTTP status mapping:

### Error Categories & Codes

| Category | Code Range | Examples |
|----------|------------|----------|
| **Configuration** | 1000-1099 | Invalid port (1001), Invalid backend URL (1002), Invalid timeouts (1004) |
| **Backend** | 1100-1199 | Backend unavailable (1100), Connection timeout (1101), No healthy backends (1104) |
| **Load Balancer** | 1200-1299 | Strategy failure (1200), Empty pool (1201), Metrics failure (1202) |
| **Health Check** | 1300-1399 | Health check failed (1313), Health check timeout (1314) |
| **Request** | 1400-1499 | Request timeout (1400), Request failed (1401), Response copy error (1402) |

### Error Context

Each error includes contextual information for debugging:

```go
// Example: Backend connection error with context
err := NewBackendConnectionError("backend-1", originalErr).
    WithContext("port", 8080).
    WithContext("attempt", 3)
```

All errors automatically map to appropriate HTTP status codes (400, 500, 502, 503, 504) for client responses.

## Key Design Patterns

- **Strategy Pattern**: Pluggable load balancing algorithms
- **Interface Segregation**: `LoadBalancingStrategy`, `MetricsProvider` interfaces
- **Structured Error Handling**: Custom error types with context and HTTP status mapping
- **Configuration Builder**: Validation with comprehensive error reporting
- **Thread Safety**: `sync.RWMutex` for backend pool, atomic counters for round-robin
- **Context Propagation**: Request timeouts and cancellation support

## Available Commands

```bash
make help            # Show all available commands
make test            # Run complete test suite (unit + integration)
make test-unit       # Run unit tests only  
make test-integration # Run integration tests only
make run-backends    # Start test backend servers
make run-lb          # Start load balancer
make kill-processes  # Clean up all running processes
make check-ports     # Check port usage status
```

## Production Readiness

This implementation includes enterprise-grade features:

- **Health checking** with configurable intervals and timeouts
- **Metrics collection** ready for Prometheus scraping
- **Structured error handling** with typed errors, context, and appropriate HTTP status codes
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
- **Rate limiting** with token bucket and sliding window algorithms
- **Weighted load balancing** with dynamic backend weighting
- **Health check diversity** (HTTP, TCP, gRPC, custom scripts)
- **Configuration hot-reloading** without service interruption
- **Admin API** for runtime backend management and configuration updates