#!/bin/bash

# Simple test script for the load balancer

echo "Testing Go Load Balancer - Step 1"
echo "=================================="
echo

# Test basic functionality
echo "1. Testing basic request forwarding..."
response=$(curl -s http://localhost:8000/)
echo "Response: $response"
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
echo "4. Testing concurrent requests..."
for i in {1..3}; do 
    curl -s http://localhost:8000/ &
done
wait
echo -e "\nConcurrent test complete!"
echo

echo "Test complete! Check the load balancer and backend logs for detailed output."
