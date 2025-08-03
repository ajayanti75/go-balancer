package strategy

import (
	"sync/atomic"

	"go-balancer/internal/pool"
)

// RoundRobinStrategy implements round-robin load balancing
type RoundRobinStrategy struct {
	counter int64
}

// NewRoundRobinStrategy creates a new round-robin strategy
func NewRoundRobinStrategy() *RoundRobinStrategy {
	return &RoundRobinStrategy{counter: 0}
}

// NextBackend returns the next backend using round-robin
func (rr *RoundRobinStrategy) NextBackend(serverPool *pool.ServerPool) *pool.Backend {
	backendCount := serverPool.GetBackendCount()
	if backendCount == 0 {
		return nil
	}

	healthyCount := serverPool.GetHealthyBackendCount()
	if healthyCount == 0 {
		return nil
	}

	// Try each backend in round-robin fashion
	// We try up to backendCount times to find a healthy backend
	for i := 0; i < backendCount; i++ {
		next := atomic.AddInt64(&rr.counter, 1)
		index := int((next - 1) % int64(backendCount))

		backend := serverPool.GetBackendByIndex(index)
		if backend != nil && backend.Healthy {
			return backend
		}
	}

	// If we get here, no healthy backends were found despite healthyCount > 0
	// This could happen due to race conditions between health checks and requests
	return nil
}

// Name returns the strategy name
func (rr *RoundRobinStrategy) Name() string {
	return "round-robin"
}
