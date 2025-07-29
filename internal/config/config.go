package config

// Config holds the configuration for our load balancer
type Config struct {
	Port     int
	Backends []string // List of backend server URLs
}
