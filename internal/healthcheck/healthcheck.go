package healthcheck

import (
	"context"
	"log"
	"net/http"
	"time"

	"go-balancer/internal/errors"
	"go-balancer/internal/pool"
)

// HealthChecker performs periodic health checks on backend servers
type HealthChecker struct {
	serverPool    *pool.ServerPool
	checkPath     string
	checkInterval time.Duration
	checkTimeout  time.Duration
	client        *http.Client
	stopCh        chan struct{}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(
	serverPool *pool.ServerPool,
	checkPath string,
	checkInterval time.Duration,
	checkTimeout time.Duration,
) *HealthChecker {
	return &HealthChecker{
		serverPool:    serverPool,
		checkPath:     checkPath,
		checkInterval: checkInterval,
		checkTimeout:  checkTimeout,
		client: &http.Client{
			Timeout: checkTimeout,
		},
		stopCh: make(chan struct{}),
	}
}

// Start begins periodic health checking
func (hc *HealthChecker) Start() {
	go hc.healthCheckLoop()
	log.Printf("Health checker started with interval %s and path %s",
		hc.checkInterval, hc.checkPath)
}

// Stop terminates health checking
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

// healthCheckLoop runs the health check at regular intervals
func (hc *HealthChecker) healthCheckLoop() {
	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	// Run an immediate check when starting
	hc.checkAllBackends()

	for {
		select {
		case <-ticker.C:
			hc.checkAllBackends()
		case <-hc.stopCh:
			log.Println("Health checker stopped")
			return
		}
	}
}

// checkAllBackends performs health checks on all backends
func (hc *HealthChecker) checkAllBackends() {
	backends := hc.serverPool.GetBackends()
	for _, backend := range backends {
		go hc.checkBackend(backend)
	}
}

// checkBackend checks the health of a single backend
func (hc *HealthChecker) checkBackend(backend *pool.Backend) {
	// Construct health check URL
	healthURL := backend.URL.String() + hc.checkPath

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), hc.checkTimeout)
	defer cancel()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		healthErr := errors.NewHealthCheckFailedError(backend.ID, err)
		log.Printf("Health check error: %v", healthErr)
		hc.serverPool.SetBackendHealth(backend.ID, false)
		return
	}

	// Add headers to identify health check requests
	req.Header.Add("User-Agent", "GoLoadBalancer-HealthCheck/1.0")

	// Perform the health check request
	resp, err := hc.client.Do(req)
	if err != nil {
		var healthErr *errors.LoadBalancerError

		// Check if it's a timeout error
		if ctx.Err() == context.DeadlineExceeded {
			healthErr = errors.NewHealthCheckTimeoutError(backend.ID)
		} else {
			healthErr = errors.NewHealthCheckFailedError(backend.ID, err)
		}

		log.Printf("Health check failed for backend %s (%s): %v",
			backend.ID, healthURL, healthErr)
		hc.serverPool.SetBackendHealth(backend.ID, false)
		return
	}
	defer resp.Body.Close()

	// Check if status code indicates health
	healthy := resp.StatusCode == http.StatusOK

	// Update backend health status if changed
	if backend.Healthy != healthy {
		if healthy {
			log.Printf("Backend %s is now healthy", backend.ID)
		} else {
			healthErr := errors.NewHealthCheckFailedError(backend.ID, nil).
				WithContext("status_code", resp.StatusCode).
				WithContext("url", healthURL)
			log.Printf("Backend %s is now unhealthy: %v", backend.ID, healthErr)
		}
		hc.serverPool.SetBackendHealth(backend.ID, healthy)
	}
}
