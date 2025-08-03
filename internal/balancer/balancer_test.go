package balancer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-balancer/internal/config"
	"go-balancer/internal/errors"
)

func TestNewLoadBalancer(t *testing.T) {
	// Test with valid configuration
	cfg := &config.Config{
		Port:                8000,
		Backends:            []string{"http://localhost:8080"},
		HealthCheckPath:     "/",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  2 * time.Second,
		BackendTimeout:      30 * time.Second,
	}

	lb, err := NewLoadBalancer(cfg)
	if err != nil {
		t.Errorf("Expected successful load balancer creation, got error: %v", err)
	}

	if lb == nil {
		t.Errorf("Expected non-nil load balancer")
	}

	// Clean up
	if lb != nil {
		lb.Stop()
	}
}

func TestNewLoadBalancerWithInvalidBackend(t *testing.T) {
	cfg := &config.Config{
		Port:                8000,
		Backends:            []string{"://invalid-url"},
		HealthCheckPath:     "/",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  2 * time.Second,
		BackendTimeout:      30 * time.Second,
	}

	lb, err := NewLoadBalancer(cfg)
	if err == nil {
		t.Errorf("Expected error for invalid backend URL")
		if lb != nil {
			lb.Stop()
		}
		return
	}

	// Check that we get a structured error
	if lbErr, ok := err.(*errors.LoadBalancerError); ok {
		if lbErr.Code != errors.ErrInvalidBackend {
			t.Errorf("Expected ErrInvalidBackend, got error code %d", lbErr.Code)
		}
	} else {
		t.Errorf("Expected LoadBalancerError, got %T", err)
	}
}

func TestLoadBalancerNoHealthyBackends(t *testing.T) {
	// Create a load balancer with a non-existent backend
	cfg := &config.Config{
		Port:                8000,
		Backends:            []string{"http://localhost:19999"}, // Non-existent port
		HealthCheckPath:     "/",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  1 * time.Second,
		BackendTimeout:      2 * time.Second,
	}

	lb, err := NewLoadBalancer(cfg)
	if err != nil {
		t.Errorf("Load balancer creation should succeed even with unreachable backend: %v", err)
		return
	}
	defer lb.Stop()

	// Create a test request
	req := httptest.NewRequest("GET", "http://localhost:8000/", nil)
	recorder := httptest.NewRecorder()

	// Make the request - should fail due to no healthy backends
	lb.ServeHTTP(recorder, req)

	// Should get service unavailable or bad gateway (both are valid for backend connection failures)
	if recorder.Code != http.StatusServiceUnavailable && recorder.Code != http.StatusBadGateway {
		t.Errorf("Expected status %d or %d, got %d", http.StatusServiceUnavailable, http.StatusBadGateway, recorder.Code)
	}

	// Check that the response body contains error information
	body := strings.ToLower(recorder.Body.String())
	expectedPhrases := []string{"backend", "connection", "failed", "unavailable"}
	found := false
	for _, phrase := range expectedPhrases {
		if strings.Contains(body, phrase) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected response to contain backend error information, got: %s", recorder.Body.String())
	}
}

func TestLoadBalancerWithMockBackend(t *testing.T) {
	// Create a mock backend server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from mock backend"))
	}))
	defer mockServer.Close()

	// Create load balancer with the mock backend
	cfg := &config.Config{
		Port:                8000,
		Backends:            []string{mockServer.URL},
		HealthCheckPath:     "/",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  2 * time.Second,
		BackendTimeout:      30 * time.Second,
	}

	lb, err := NewLoadBalancer(cfg)
	if err != nil {
		t.Errorf("Load balancer creation failed: %v", err)
		return
	}
	defer lb.Stop()

	// Give health checker time to mark backend as healthy
	time.Sleep(100 * time.Millisecond)

	// Create a test request
	req := httptest.NewRequest("GET", "http://localhost:8000/test", nil)
	recorder := httptest.NewRecorder()

	// Make the request
	lb.ServeHTTP(recorder, req)

	// Should get success
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	// Check response body
	expected := "Hello from mock backend"
	if recorder.Body.String() != expected {
		t.Errorf("Expected response '%s', got '%s'", expected, recorder.Body.String())
	}
}

func TestLoadBalancerRoundRobin(t *testing.T) {
	// Create multiple mock backend servers
	backends := make([]*httptest.Server, 3)
	backendURLs := make([]string, 3)

	for i := 0; i < 3; i++ {
		port := i
		backends[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Backend %d", port)))
		}))
		backendURLs[i] = backends[i].URL
		defer backends[i].Close()
	}

	// Create load balancer with multiple backends
	cfg := &config.Config{
		Port:                8000,
		Backends:            backendURLs,
		HealthCheckPath:     "/",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  2 * time.Second,
		BackendTimeout:      30 * time.Second,
	}

	lb, err := NewLoadBalancer(cfg)
	if err != nil {
		t.Errorf("Load balancer creation failed: %v", err)
		return
	}
	defer lb.Stop()

	// Give health checker time to mark backends as healthy
	time.Sleep(100 * time.Millisecond)

	// Make multiple requests and verify round-robin distribution
	responses := make(map[string]int)
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("GET", "http://localhost:8000/", nil)
		recorder := httptest.NewRecorder()

		lb.ServeHTTP(recorder, req)

		if recorder.Code == http.StatusOK {
			responses[recorder.Body.String()]++
		}
	}

	// Each backend should have received 2 requests
	expectedCount := 2
	for response, count := range responses {
		if count != expectedCount {
			t.Errorf("Expected %d requests for response '%s', got %d", expectedCount, response, count)
		}
	}
}
