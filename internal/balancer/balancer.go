package balancer

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"go-balancer/internal/config"
	"go-balancer/internal/errors"
	"go-balancer/internal/healthcheck"
	"go-balancer/internal/metrics"
	"go-balancer/internal/pool"
	"go-balancer/internal/strategy"
)

// LoadBalancer represents our load balancer
type LoadBalancer struct {
	config        *config.Config
	client        *http.Client
	serverPool    *pool.ServerPool
	strategy      strategy.LoadBalancingStrategy
	healthChecker *healthcheck.HealthChecker
	metrics       *metrics.Metrics

	metricsProvider metrics.MetricsProvider
}

// NewLoadBalancer creates a new LoadBalancer instance
func NewLoadBalancer(cfg *config.Config) (*LoadBalancer, error) {
	serverPool := pool.NewServerPool()

	// Add all configured backends to the pool
	for _, backend := range cfg.Backends {
		if err := serverPool.AddBackend(backend); err != nil {
			return nil, errors.NewInvalidBackendError(backend, err)
		}
	}

	// Validate we have at least one backend
	if serverPool.GetBackendCount() == 0 {
		return nil, errors.NewPoolEmptyError()
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

	m := metrics.NewMetrics()
	return &LoadBalancer{
		config:          cfg,
		client:          &http.Client{},
		serverPool:      serverPool,
		strategy:        strategy.NewRoundRobinStrategy(),
		healthChecker:   healthChecker,
		metrics:         m,
		metricsProvider: metrics.NewPrometheusMetricsProvider(m),
	}, nil
}

// getNextHealthyBackend uses the configured strategy to get next backend
func (lb *LoadBalancer) getNextHealthyBackend() (*pool.Backend, error) {
	backend := lb.strategy.NextBackend(lb.serverPool)
	if backend == nil {
		healthyCount := lb.serverPool.GetHealthyBackendCount()
		totalCount := lb.serverPool.GetBackendCount()

		if totalCount == 0 {
			return nil, errors.NewPoolEmptyError()
		}

		return nil, errors.NewNoHealthyBackendsError().
			WithContext("healthy_count", healthyCount).
			WithContext("total_count", totalCount)
	}
	return backend, nil
}

// ServeHTTP implements the http.Handler interface
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get next healthy backend using round-robin
	backend, err := lb.getNextHealthyBackend()
	if err != nil {
		log.Printf("Failed to get healthy backend: %v", err)

		// Convert structured error to appropriate HTTP response
		if lbErr, ok := err.(*errors.LoadBalancerError); ok {
			http.Error(w, lbErr.Message, lbErr.HTTPStatusCode())
		} else {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		}
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
		reqErr := errors.NewRequestFailedError(err).WithContext("backend", backend.ID)
		http.Error(w, reqErr.Message, reqErr.HTTPStatusCode())
		return
	}

	// Copy headers from original request
	backendReq.Header = r.Header.Clone()

	// Copy query parameters
	backendReq.URL.RawQuery = r.URL.RawQuery

	// Make the request to the backend server
	start := time.Now()
	resp, err := lb.client.Do(backendReq)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Error forwarding request to backend %s: %v", backend.ID, err)

		// Determine the type of error
		var lbErr *errors.LoadBalancerError
		if ctx.Err() == context.DeadlineExceeded {
			lbErr = errors.NewBackendTimeoutError(backend.ID, err)
		} else {
			lbErr = errors.NewBackendConnectionError(backend.ID, err)
		}

		// Record failure in metrics
		lb.metrics.RecordFailure(backend.ID)

		// Mark backend as unhealthy for future requests
		lb.serverPool.SetBackendHealth(backend.ID, false)

		http.Error(w, lbErr.Message, lbErr.HTTPStatusCode())
		return
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode >= 500 {
		log.Printf("Backend %s returned error status: %d", backend.ID, resp.StatusCode)

		respErr := errors.NewBackendResponseError(backend.ID, resp.StatusCode)
		lb.metrics.RecordFailure(backend.ID)

		// Don't mark backend as unhealthy for 5xx errors - might be temporary
		// Only health checks should determine backend health

		http.Error(w, respErr.Message, respErr.HTTPStatusCode())
		return
	}

	// Record successful request in metrics
	lb.metrics.RecordRequest(backend.ID, duration)

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
		// Note: We can't change status code after WriteHeader, but we can log the error
		copyErr := errors.NewResponseCopyError(err).WithContext("backend", backend.ID)
		log.Printf("Response copy error: %v", copyErr)
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

// GetMetricsProvider returns the metrics provider
func (lb *LoadBalancer) GetMetricsProvider() metrics.MetricsProvider {
	return lb.metricsProvider
}
