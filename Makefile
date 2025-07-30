# Go Load Balancer Makefile

.PHONY: help build run-backend run-lb test clean kill-processes check-ports

# Default target
help:
	@echo "Go Load Balancer - Available commands:"
	@echo "  build         - Build the load balancer binary"
	@echo "  run-backend   - Run the test backend server on port 8080"
	@echo "  run-lb        - Run the load balancer on port 8000"
	@echo "  test          - Run test script (starts/stops processes automatically)"
	@echo "  demo          - Run complete demo (automated test)"
	@echo "  clean         - Clean build artifacts"
	@echo "  kill-processes - Kill any running Go processes"
	@echo "  check-ports   - Check what's running on our ports"

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
	@# Kill any process using our specific ports
	@lsof -ti :8000 | xargs -r kill -9 || echo "Port 8000 is free"
	@lsof -ti :8080 | xargs -r kill -9 || echo "Port 8080 is free"
	@lsof -ti :8081 | xargs -r kill -9 || echo "Port 8081 is free"
	@lsof -ti :8082 | xargs -r kill -9 || echo "Port 8082 is free"
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

# Run automated test
test: kill-processes
	@echo "Running automated test..."
	@echo "Starting multiple test backend servers..."
	@go run test/test_backend.go -num=3 -ports="8080,8081,8082" &
	@sleep 3
	@echo "Starting load balancer..."
	@go run main.go -port=8000 -backends="http://localhost:8080,http://localhost:8081,http://localhost:8082" > /tmp/lb.log 2>&1 &
	@sleep 5
	@echo "Running tests..."
	@./test.sh
	@echo "Load balancer logs:"
	@cat /tmp/lb.log
	@echo "Test completed. Cleaning up..."
	@$(MAKE) kill-processes

# Run complete demo
demo: test

# Clean build artifacts
clean: kill-processes
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
