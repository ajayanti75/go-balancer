package config

import "time"

// Config holds the configuration for our load balancer
type Config struct {
	Port                int
	Backends            []string      // List of backend server URLs
	HealthCheckPath     string        // Path to use for health checks
	HealthCheckInterval time.Duration // Interval between health checks
	HealthCheckTimeout  time.Duration // Timeout for health check requests
	BackendTimeout      time.Duration // Timeout for backend requests
}
