# Go Load Balancer Makefile

.PHONY: help build run-backend run-lb test test-unit test-integration clean kill-processes check-ports

# Default target
help:
	@echo "Go Load Balancer - Available commands:"
	@echo "  build            - Build the load balancer binary"
	@echo "  test             - Run complete test suite (unit + integration)"
	@echo "  test-unit        - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  run-backends     - Start test backend servers"
	@echo "  run-lb           - Run the load balancer"
	@echo "  demo             - Run complete demo (automated integration test)"
	@echo "  clean            - Clean build artifacts"
	@echo "  kill-processes   - Kill any running Go processes"
	@echo "  check-ports      - Check what's running on our ports"

# Build the load balancer
build:
	@echo "Building load balancer..."
	go build -o bin/lb main.go
	go build -o bin/test_backend test/test_backend.go

# Check what's running on our ports
check-ports:
	@echo "Checking ports 8000, 8080, 8081, 8082..."
	@lsof -i :8000 -i :8080 -i :8081 -i :8082 || echo "Ports are free"

# Kill any processes that might be using our ports
kill-processes:
	@echo "Killing any running processes on our ports..."
	@# Kill Go processes first
	@pkill -f "go run" || echo "No go run processes found"
	@pkill -f "bin/lb" || echo "No lb processes found"
	@pkill -f "bin/backend" || echo "No backend processes found"
	@# Kill any process using our specific ports (macOS compatible)
	@lsof -ti :8000 | while read pid; do kill -9 $$pid 2>/dev/null || true; done || echo "Port 8000 is free"
	@lsof -ti :8080 | while read pid; do kill -9 $$pid 2>/dev/null || true; done || echo "Port 8080 is free"
	@lsof -ti :8081 | while read pid; do kill -9 $$pid 2>/dev/null || true; done || echo "Port 8081 is free"
	@lsof -ti :8082 | while read pid; do kill -9 $$pid 2>/dev/null || true; done || echo "Port 8082 is free"
	@# Give processes time to clean up
	@sleep 2
	@echo "Port cleanup complete"

# Run backend servers (multiple)
run-backends: kill-processes
	@echo "Starting multiple test backend servers..."
	go run test/test_backend.go -num=3 -ports="8080,8081,8082"

# Run load balancer with multiple backends
run-lb: kill-processes
	@echo "Starting load balancer on port 8000 with multiple backends..."
	go run main.go -port=8000 -backends="http://localhost:8080,http://localhost:8081,http://localhost:8082"

# Run complete test suite
test: test-unit test-integration

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v ./internal/...

# Run integration tests
test-integration: kill-processes
	@echo "Running integration tests..."
	@echo "Starting multiple test backend servers..."
	@go run test/test_backend.go -num=3 -ports="8080,8081,8082" &
	@sleep 3
	@echo "Starting load balancer..."
	@go run main.go -port=8000 -backends="http://localhost:8080,http://localhost:8081,http://localhost:8082" > /tmp/lb.log 2>&1 &
	@sleep 5
	@echo "Running integration test script..."
	@echo "Testing Go Load Balancer - Round-Robin Distribution"
	@echo "=================================================="
	@echo "1. Testing round-robin distribution (6 requests)..."
	@for i in 1 2 3 4 5 6; do \
		response=$$(curl -s http://localhost:8000/); \
		echo "Request $$i: $$response"; \
	done
	@echo "2. Testing with path /api/test..."
	@response=$$(curl -s http://localhost:8000/api/test); echo "Response: $$response"
	@echo "3. Testing with custom headers..."
	@response=$$(curl -s -H "X-Test-Header: test-value" http://localhost:8000/); echo "Response: $$response"
	@echo "4. Testing concurrent requests (should see different backends)..."
	@for i in 1 2 3 4 5 6; do curl -s http://localhost:8000/ & done; wait
	@echo "Concurrent test complete!"
	@echo "Load balancer logs:"
	@cat /tmp/lb.log
	@echo "Integration test completed. Cleaning up..."
	@$(MAKE) kill-processes

# Run complete demo
demo: test-integration

# Clean build artifacts
clean: kill-processes
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
