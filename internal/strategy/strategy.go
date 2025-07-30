package strategy

import "go-balancer/internal/pool"

// LoadBalancingStrategy defines different load balancing algorithms
type LoadBalancingStrategy interface {
	NextBackend(serverPool *pool.ServerPool) *pool.Backend
	Name() string
}
