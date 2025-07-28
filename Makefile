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
	go build -o bin/backend cmd/backend/main.go

# Check what's running on our ports
check-ports:
	@echo "Checking ports 8000 and 8080..."
	@lsof -i :8000 -i :8080 || echo "Ports are free"

# Kill any Go processes that might be running
kill-processes:
	@echo "Killing any running Go processes..."
	@pkill -f "go run" || echo "No Go processes found"
	@pkill -f "bin/lb" || echo "No lb processes found"
	@pkill -f "bin/backend" || echo "No backend processes found"
	@sleep 1

# Run backend server
run-backend: kill-processes
	@echo "Starting backend server on port 8080..."
	go run cmd/backend/main.go

# Run load balancer on port 8000
run-lb: kill-processes
	@echo "Starting load balancer on port 8000..."
	go run main.go -port=8000

# Run automated test
test: kill-processes
	@echo "Running automated test..."
	@echo "Starting backend server..."
	@go run cmd/backend/main.go &
	@echo "Waiting for backend to start..."
	@sleep 2
	@echo "Starting load balancer..."
	@go run main.go -port=8000 &
	@echo "Waiting for load balancer to start..."
	@sleep 2
	@echo "Running tests..."
	@./test.sh
	@echo "Stopping processes..."
	@pkill -f "go run" || echo "Processes already stopped"

# Run complete demo
demo: test

# Clean build artifacts
clean: kill-processes
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
