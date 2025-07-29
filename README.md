# Go Load Balancer

A production-ready HTTP load balancer implementation in Go, built as part of the [Coding Challenges Load Balancer Challenge](https://codingchallenges.fyi/challenges/challenge-load-balancer/).

## Project Overview

This project implements a robust HTTP load balancer with:
- ✅ **Round-robin load balancing** across multiple backend servers
- ✅ **Concurrent request handling** using goroutines
- ✅ **Thread-safe operations** with proper mutex usage
- ✅ **Graceful error handling** and backend failure detection
- ✅ **Modular architecture** with clean separation of concerns
- 🚧 **Health checking** (Step 3 - Coming Soon)
- 🚧 **Docker containerization** (Step 4 - Final production deployment)

## Project Structure

```
go-balancer/
├── main.go                    # Entry point for the load balancer
├── internal/                  # Core load balancer packages
│   ├── config/               # Configuration management
│   │   └── config.go
│   ├── pool/                 # Backend server pool management
│   │   └── server_pool.go
│   └── balancer/             # Load balancing logic
│       └── balancer.go
├── test/                     # Testing infrastructure (separate from main app)
│   └── test_backend.go       # Mock backend servers for testing
├── test.sh                   # Automated test script
├── Makefile                  # Build and test automation
├── go.mod
└── README.md
```

## Implementation Status

### ✅ Step 1: Basic HTTP Server and Request Forwarding (COMPLETED)
- [x] HTTP server listening on configurable port
- [x] Detailed request logging with client information
- [x] Request forwarding to backend servers
- [x] Concurrent request handling with goroutines
- [x] Comprehensive error handling and logging
- [x] Command-line configuration support

### ✅ Step 2: Round-Robin Load Balancing (COMPLETED)
- [x] Support for multiple backend servers (3 by default)
- [x] Thread-safe round-robin algorithm using atomic counters
- [x] Even distribution of requests across healthy backends
- [x] Modular architecture with separate packages:
  - **Config**: Centralized configuration management
  - **Pool**: Backend server pool with thread-safe operations
  - **Balancer**: Load balancing logic and HTTP request handling
- [x] Dynamic backend management (add/remove backends)
- [x] Backend failure detection and automatic marking as unhealthy

### 🚧 Step 3: Health Checking (TODO)
- [ ] Periodic health checks for backend servers
- [ ] Automatic removal of unhealthy servers from rotation
- [ ] Re-addition of servers when they become healthy again
- [ ] Configurable health check intervals and timeouts

### 🚧 Step 4: Docker Containerization & Integration Testing (TODO)
- [ ] **Multi-stage Dockerfile** for optimized production builds
- [ ] **Docker Compose setup** with load balancer + multiple backend services
- [ ] **Container networking** configuration for service discovery
- [ ] **Environment-based configuration** (ports, backend URLs, health check settings)
- [ ] **Integration test suite** running in containerized environment
- [ ] **End-to-end testing** with real Docker containers
- [ ] **Production deployment** documentation with Docker best practices
- [ ] **Container health checks** and monitoring setup
- [ ] **Volume mounting** for configuration and logs
- [ ] **Multi-architecture builds** (AMD64, ARM64) for deployment flexibility

## How to Run

### Prerequisites
- Go 1.21 or later installed
- For Step 4 (future): Docker and Docker Compose

### Running the Load Balancer

1. **Start test backend servers** (for testing):
   ```bash
   # Terminal 1: Start 3 test backend servers on ports 8080, 8081, 8082
   make run-backends
   ```

2. **Run the load balancer**:
   ```bash
   # Terminal 2: Run the load balancer on port 8000
   make run-lb
   ```

3. **Test the setup**:
   ```bash
   # Terminal 3: Send test requests
   curl http://localhost:8000/
   
   # Or run our comprehensive automated test:
   make test
   ```

### Available Commands

```bash
make help           # Show all available commands
make build          # Build binaries (lb and test_backend)
make run-backends   # Start 3 test backend servers
make run-lb         # Start load balancer on port 8000
make test           # Run comprehensive automated tests
make demo           # Same as test
make clean          # Clean up build artifacts
make kill-processes # Kill any running processes and free ports
make check-ports    # Check what's using our ports (8000-8082)
```

## Configuration

The load balancer supports the following configuration options:

- `--port`: Port for load balancer to listen on (default: 8000)
- `--backends`: Comma-separated list of backend URLs (default: 3 backends on localhost)

### Example Usage

```bash
# Run with default settings (3 backends: 8080, 8081, 8082)
go run main.go

# Run on custom port with custom backends
go run main.go -port=9000 -backends="http://localhost:3000,http://localhost:3001"

# Run test backends on custom ports
go run test/test_backend.go -num=2 -ports="3000,3001"
```

## Test Coverage & Real-World Value

Our test suite validates several critical load balancer capabilities:

### 1. **Round-Robin Distribution Testing**
```bash
# Tests even distribution across backends
for i in {1..6}; do curl http://localhost:8000/; done
```
**Real-World Value**: Ensures no single backend gets overwhelmed while others sit idle.

### 2. **Path Forwarding Testing**
```bash
# Tests different URL paths
curl http://localhost:8000/api/test
```
**Real-World Value**: Critical for microservices where different paths may have different performance characteristics. Load balancers must preserve the full request path, query parameters, and routing context.

### 3. **Header Forwarding Testing**
```bash
# Tests custom header propagation
curl -H "X-Test-Header: test-value" http://localhost:8000/
```
**Real-World Value**: 
- **Authentication**: JWT tokens, API keys in headers must reach backends
- **Tracing**: Distributed tracing headers (X-Trace-ID, X-Request-ID) for observability
- **Content Negotiation**: Accept, Content-Type headers for API versioning
- **Load Balancer Metadata**: Custom headers like X-Forwarded-For, X-Real-IP

### 4. **Concurrent Request Testing**
```bash
# Tests thread safety under load
for i in {1..6}; do curl http://localhost:8000/ & done; wait
```
**Real-World Value**: Validates that the load balancer can handle multiple simultaneous requests without race conditions, request mixing, or crashes - essential for production traffic.

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
