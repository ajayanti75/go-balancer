# Go Load Balancer

A simple HTTP load balancer implementation in Go, built as part of the [Coding Challenges Load Balancer Challenge](https://codingchallenges.fyi/challenges/challenge-load-balancer/).

## Project Overview

This project implements a basic HTTP load balancer that:
- Distributes client requests across multiple backend servers
- Performs health checks on backend servers
- Uses round-robin load balancing algorithm
- Handles server failures gracefully

## Project Structure

```
go-balancer/
├── main.go           # Entry point for the load balancer
├── cmd/
│   └── backend/      # Simple backend server for testing
│       └── main.go
├── test.sh           # Automated test script
├── Makefile         # Build and test automation
├── go.mod
└── README.md
```

## Current Implementation Status

### ✅ Step 1: Basic HTTP Server and Request Forwarding (COMPLETED)
- [x] Create HTTP server that listens on specified port
- [x] Log incoming connections with client details
- [x] Forward requests to a single backend server
- [x] Handle concurrent requests using goroutines
- [x] Proper error handling and logging
- [x] Command-line configuration support

### 🚧 Step 2: Round-Robin Load Balancing (TODO)
- [ ] Support multiple backend servers
- [ ] Implement round-robin algorithm
- [ ] Distribute requests evenly across servers

### 🚧 Step 3: Health Checking (TODO)
- [ ] Periodic health checks for backend servers
- [ ] Remove unhealthy servers from rotation
- [ ] Re-add servers when they become healthy again

## How to Run

### Prerequisites
- Go 1.21 or later installed

### Running the Load Balancer

1. **Start a backend server** (for testing):
   ```bash
   # Terminal 1: Start our Go backend server
   make run-backend
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
   
   # Or run our automated test script:
   make test
   ```

### Available Commands

```bash
make help          # Show all available commands
make build         # Build binaries
make run-backend   # Start backend server on port 8080
make run-lb        # Start load balancer on port 8000
make test          # Run automated tests (handles start/stop)
make demo          # Same as test
make clean         # Clean up build artifacts
make kill-processes # Kill any running processes
make check-ports   # Check what's using ports 8000/8080
```

## Configuration

The load balancer currently supports the following configuration (via command line flags):

- `--port`: Port to listen on (default: 80)
- `--backend`: Backend server URL (default: http://localhost:8080)

### Example Usage

```bash
# Run on custom port with custom backend
go run main.go -port=9000 -backend=http://localhost:3000

# Run backend on custom port
go run cmd/backend/main.go -port=3000
```

## Expected Output

When you run a test request, you should see:
- Load balancer logs showing the incoming request details
- Backend server logs showing the forwarded request
- Response from the backend server returned to the client

Example:
```
# Load balancer logs:
2025/07/28 16:23:00 Load balancer starting on port 8000
2025/07/28 16:23:00 Forwarding requests to: http://localhost:8080
2025/07/28 16:23:00 Received GET request on / from [::1]:55102:
2025/07/28 16:23:00 Host: localhost:8000
2025/07/28 16:23:00 Response from server: 200 OK

# Backend server logs:
2025/07/28 16:22:57 Backend server starting on port 8080
2025/07/28 16:23:00 Received request from [::1]:55103: GET /
2025/07/28 16:23:00 Replied with a hello message

# Client response:
Hello From Backend Server on port 8080
```
