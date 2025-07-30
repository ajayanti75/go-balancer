# Go Load Balancer

A production-ready HTTP load balancer implementation in Go, built following the [Coding Challenges Load Balancer Challenge](https://codingchallenges.fyi/challenges/challenge-load-balancer/).

## Features

- ✅ **Round-robin load balancing** across multiple backends
- ✅ **Health checking** with automatic failure detection and recovery
- ✅ **Concurrent request handling** using goroutines
- ✅ **Thread-safe operations** with proper mutex usage
- ✅ **Configurable timeouts** for health checks and backend requests
- ✅ **Graceful error handling** and comprehensive logging
- 🚧 **Docker containerization** (Step 4 - Final production deployment)

## Project Structure

```
go-balancer/
├── main.go                    # Entry point
├── internal/                  # Core packages
│   ├── config/               # Configuration management
│   ├── pool/                 # Backend server pool management
│   ├── balancer/             # Load balancing logic
│   └── healthcheck/          # Health checking system
├── test/test_backend.go      # Testing infrastructure
├── test.sh                   # Automated test script
└── Makefile                  # Build and test automation
```

## Implementation Status

### ✅ Steps 1-3: Completed
- **Basic HTTP forwarding** with detailed logging and concurrent handling
- **Round-robin load balancing** with thread-safe atomic operations
- **Health checking** with periodic monitoring and automatic failure recovery
- **Modular architecture** with clean separation of concerns (config, pool, balancer, healthcheck)
- **Configurable timeouts** for both health checks and backend requests
- **Production-ready features**: Dynamic backend management, error handling, comprehensive logging

### 🚧 Step 4: Docker Containerization & Integration Testing
Next phase will include:
- Multi-stage Dockerfile for optimized production builds
- Docker Compose setup with load balancer + multiple backend services
- Containerized integration test suite with automated testing
- Production deployment documentation and best practices

## Quick Start

### Run with default settings:
```bash
make test                    # Automated test with 3 backends
```

### Run manually:
```bash
# Start backends
go run test/test_backend.go -num=3 -ports="8080,8081,8082"

# Start load balancer  
go run main.go

# Test requests
curl http://localhost:8000/
```

## Configuration

Customize behavior with command-line flags:
```bash
go run main.go \
  -port=8000 \
  -backends="http://localhost:8080,http://localhost:8081,http://localhost:8082" \
  -health-interval=10 \
  -health-timeout=2 \
  -backend-timeout=30 \
  -health-path="/"
```

## Architecture Highlights

- **Health Checking**: Periodic monitoring with automatic failure detection and recovery
- **Thread Safety**: RWMutex for backend pool, atomic counters for round-robin selection
- **Context Timeouts**: Configurable timeouts for health checks (2s) and backend requests (30s)
- **Error Handling**: Backends marked unhealthy on failures, graceful degradation
- **Testing**: Comprehensive test suite including failure/recovery scenarios

## Available Commands

```bash
make help           # Show all commands
make test           # Run full automated test suite  
make kill-processes # Clean up all processes
make check-ports    # Check port usage
```

## Future Improvements

### Phase 1: Production Readiness
**1. Interfaces & Testability**
- Add interfaces for LoadBalancer, BackendPool, HealthChecker
- Enable dependency injection and mocking for unit tests
- Decouple components for better testability

**2. Structured Error Handling**
```go
type LoadBalancerError struct {
    Op      string // Operation that failed
    Backend string // Backend that caused error
    Err     error  // Underlying error
}
```

**3. Configuration Validation**
- Builder pattern for configuration with validation
- Validate ports (1-65535), URLs, timeout relationships
- Sensible defaults with override capabilities

**4. Unit Testing**
- Comprehensive test suite for individual components
- Mock HTTP clients and dependencies
- Test concurrent access patterns

### Phase 2: Advanced Features
**5. Load Balancing Strategies**
```go
type LoadBalancingStrategy interface {
    NextBackend(pool BackendPool) *Backend
    Name() string
}
```
- Round Robin (current)
- Weighted Round Robin
- Least Connections
- IP Hash (sticky sessions)
- Random selection

**6. Observability & Metrics**
```go
type Metrics struct {
    TotalRequests     int64
    SuccessfulRequests int64
    BackendLatencies  map[string][]time.Duration
    HealthCheckPasses map[string]int64
    ResponseTimes     []time.Duration
}
```

**7. Circuit Breaker Pattern**
- Prevent cascading failures
- Automatic recovery detection
- Configurable failure thresholds

### Phase 3: Enterprise Features
**8. Request Tracing**
- Distributed tracing with correlation IDs
- Request lifecycle tracking
- Performance bottleneck identification

**9. Rate Limiting**
- Per-client request limiting
- Backend protection from overload
- Configurable limits and windows

**10. Admin API**
```go
GET  /admin/health      # Overall system health
GET  /admin/backends    # Backend status
POST /admin/backends    # Add backend
PUT  /admin/config      # Runtime configuration
```

**11. Connection Pooling**
- Reuse HTTP connections to backends
- Configurable pool sizes
- Connection lifecycle management

This load balancer demonstrates production-ready Go concurrency patterns, proper error handling, and enterprise-grade health checking capabilities.

## Expected Output

When running tests, you should see perfect round-robin distribution:
- Load balancer logs showing the incoming request details
- Backend server logs showing the forwarded request
- Response from the backend server returned to the client

```
# Load balancer logs:
2025/07/29 10:29:57 Load balancer starting on port 8000
2025/07/29 10:29:57 Forwarding requests to backends: [http://localhost:8080 http://localhost:8081 http://localhost:8082]
2025/07/29 10:29:58 Received GET request on / from [::1]:50292
2025/07/29 10:29:58 Forwarding to backend: backend-1 (http://localhost:8080)
2025/07/29 10:29:58 Response from backend backend-1: 200 OK

# Test backend logs:
2025/07/29 10:29:54 Test backend server starting on port 8080
2025/07/29 10:29:58 Received request from [::1]:50293: GET /
2025/07/29 10:29:58 Replied with a hello message

# Client responses show perfect round-robin:
Request 1: Hello From Backend Server on port 8080
Request 2: Hello From Backend Server on port 8081  
Request 3: Hello From Backend Server on port 8082
Request 4: Hello From Backend Server on port 8080  # Back to first
Request 5: Hello From Backend Server on port 8081  # Second
Request 6: Hello From Backend Server on port 8082  # Third
```

## Future: Docker Containerization (Step 4)

Once health checking is implemented, the final step will be containerizing the entire application for production deployment. This will include:

### Planned Docker Architecture
```
go-balancer-docker/
├── Dockerfile                    # Multi-stage build for load balancer
├── docker-compose.yml           # Orchestrate LB + multiple backends
├── docker-compose.test.yml      # Integration testing environment
├── backend.Dockerfile           # Lightweight backend service container
├── config/
│   ├── lb-config.yml           # Load balancer configuration
│   └── backend-config.yml      # Backend service configuration
└── tests/
    ├── integration/            # Docker-based integration tests
    │   ├── test-suite.sh      # End-to-end test automation
    │   └── load-test.js       # Performance testing with containerized setup
    └── docker-test.yml        # Test-specific compose configuration
```

### Key Docker Features (Planned)
- **Multi-stage Dockerfile**: Optimize build size with separate build and runtime stages
- **Docker Compose**: Orchestrate load balancer + 3 backend services with proper networking
- **Environment Configuration**: Use environment variables for all configuration (ports, URLs, timeouts)
- **Health Checks**: Docker container health checks for both LB and backend services
- **Integration Tests**: Full end-to-end testing in containerized environment
- **Production Ready**: Volume mounts, logging, monitoring, and deployment best practices

### Planned Commands (Future)
```bash
# Build all containers
docker-compose build

# Run entire stack (LB + 3 backends)
docker-compose up -d

# Run integration tests in containers
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Scale backend services
docker-compose up --scale backend=5

# Load testing against containerized setup
docker-compose exec test npm run load-test
```

This containerized setup will provide a production-ready deployment option with proper isolation, scalability, and testing capabilities.

## Architecture Notes

### Thread Safety
- Uses `sync.RWMutex` for safe concurrent access to backend pool
- Atomic counters for lock-free round-robin selection
- Goroutines handle each request independently

### Error Handling
- Backends marked unhealthy on connection failures
- Graceful degradation when backends become unavailable
- Comprehensive logging for debugging and monitoring

### Testing Infrastructure
The `test/` directory contains testing utilities separate from the main application:
- **test_backend.go**: Mock backend servers that simulate real application servers
- **test.sh**: Comprehensive test scenarios including edge cases
- **Makefile**: Automated testing with proper process management

This separation ensures testing code doesn't pollute the production load balancer codebase.
