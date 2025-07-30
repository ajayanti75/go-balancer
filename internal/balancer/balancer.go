package balancer

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"

	"go-balancer/internal/config"
	"go-balancer/internal/healthcheck"
	"go-balancer/internal/pool"
)

// LoadBalancer represents our load balancer
type LoadBalancer struct {
	config        *config.Config
	client        *http.Client
	serverPool    *pool.ServerPool
	current       int64 // atomic counter for round-robin
	healthChecker *healthcheck.HealthChecker
}

// NewLoadBalancer creates a new LoadBalancer instance
func NewLoadBalancer(cfg *config.Config) (*LoadBalancer, error) {
	serverPool := pool.NewServerPool()

	// Add all configured backends to the pool
	for _, backend := range cfg.Backends {
		if err := serverPool.AddBackend(backend); err != nil {
			return nil, fmt.Errorf("failed to add backend %s: %w", backend, err)
		}
	}

	// Create health checker
	healthChecker := healthcheck.NewHealthChecker(
		serverPool,
		cfg.HealthCheckPath,
		cfg.HealthCheckInterval,
		cfg.HealthCheckTimeout,
	)

	// Start health checks
	healthChecker.Start()

	return &LoadBalancer{
		config:        cfg,
		client:        &http.Client{},
		serverPool:    serverPool,
		current:       0,
		healthChecker: healthChecker,
	}, nil
}

// getNextHealthyBackend implements round-robin among healthy backends
func (lb *LoadBalancer) getNextHealthyBackend() *pool.Backend {
	backendCount := lb.serverPool.GetBackendCount()
	if backendCount == 0 {
		return nil
	}

	// Try each backend in round-robin fashion
	for i := 0; i < backendCount; i++ {
		next := atomic.AddInt64(&lb.current, 1)
		index := int((next - 1) % int64(backendCount))

		backend := lb.serverPool.GetBackendByIndex(index)
		if backend != nil && backend.Healthy {
			return backend
		}
	}

	// No healthy backends found
	return nil
}

// ServeHTTP implements the http.Handler interface
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get next healthy backend using round-robin
	backend := lb.getNextHealthyBackend()
	if backend == nil {
		log.Printf("No healthy backends available")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	log.Printf("Received %s request on %s from %s:",
		r.Method, r.URL.Path, r.RemoteAddr)
	log.Printf("Host: %s", r.Host)
	log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))
	log.Printf("Forwarding to backend: %s (%s)", backend.ID, backend.URL.String())

	// Create context with timeout for the backend request
	ctx, cancel := context.WithTimeout(r.Context(), lb.config.BackendTimeout)
	defer cancel()

	// Create a new request to forward to the selected backend
	backendReq, err := http.NewRequestWithContext(ctx, r.Method, backend.URL.String()+r.URL.Path, r.Body)
	if err != nil {
		log.Printf("Error creating backend request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers from original request
	backendReq.Header = r.Header.Clone()

	// Copy query parameters
	backendReq.URL.RawQuery = r.URL.RawQuery

	// Make the request to the backend server
	resp, err := lb.client.Do(backendReq)
	if err != nil {
		log.Printf("Error forwarding request to backend %s: %v", backend.ID, err)

		// Mark backend as unhealthy for future requests
		lb.serverPool.SetBackendHealth(backend.ID, false)

		http.Error(w, "Backend Server Error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Log the response from backend
	log.Printf("Response from backend %s: %s", backend.ID, resp.Status)

	// Copy response headers back to client
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set the status code
	w.WriteHeader(resp.StatusCode)

	// Copy the response body back to client
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

// AddBackend dynamically adds a new backend server
func (lb *LoadBalancer) AddBackend(backendURL string) error {
	return lb.serverPool.AddBackend(backendURL)
}

// RemoveBackend dynamically removes a backend server
func (lb *LoadBalancer) RemoveBackend(id string) bool {
	return lb.serverPool.RemoveBackend(id)
}

// GetBackends returns current backend status
func (lb *LoadBalancer) GetBackends() []*pool.Backend {
	return lb.serverPool.GetBackends()
}

// Stop gracefully shuts down the load balancer
func (lb *LoadBalancer) Stop() {
	if lb.healthChecker != nil {
		lb.healthChecker.Stop()
	}
}
