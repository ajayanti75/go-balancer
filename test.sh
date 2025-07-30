#!/bin/bash

# Simple test script for the load balancer

echo "Testing Go Load Balancer - Step 2: Round-Robin"
echo "=============================================="
echo

# Test round-robin distribution
echo "1. Testing round-robin distribution (6 requests)..."
echo "Expecting requests to be distributed across ports 8080, 8081, 8082"
for i in {1..6}; do
    response=$(curl -s http://localhost:8000/)
    echo "Request $i: $response"
done
echo

# Test with specific path
echo "2. Testing with path /api/test..."
response=$(curl -s http://localhost:8000/api/test)
echo "Response: $response"
echo

# Test with headers
echo "3. Testing with custom headers..."
response=$(curl -s -H "X-Test-Header: test-value" http://localhost:8000/)
echo "Response: $response"
echo

# Test concurrent requests
echo "4. Testing concurrent requests (should see different backends)..."
for i in {1..6}; do 
    curl -s http://localhost:8000/ &
done
wait
echo -e "\nConcurrent test complete!"
echo

# Test health check functionality
echo "5. Testing health check recovery..."
echo "Note: Health check test disabled in automated mode due to timing complexity"
echo "To test health checking manually:"
echo "  1. Run: make run-backends (in terminal 1)"
echo "  2. Run: make run-lb (in terminal 2)" 
echo "  3. Kill one backend process and observe recovery"
echo

echo "Test complete! Check the load balancer logs to verify:"
echo "1. Round-robin distribution"
echo "2. Health check detection of failed backend"
echo "3. Backend recovery and re-inclusion in rotation"
