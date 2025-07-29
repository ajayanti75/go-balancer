package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"go-balancer/internal/balancer"
	"go-balancer/internal/config"
)

func main() {
	// Parse command line flags
	var (
		port     = flag.Int("port", 8000, "Port to listen on")
		backends = flag.String("backends", "http://localhost:8080,http://localhost:8081,http://localhost:8082", "Comma-separated list of backend servers")
	)
	flag.Parse()

	// Parse backends string into slice
	backendList := strings.Split(*backends, ",")
	for i, backend := range backendList {
		backendList[i] = strings.TrimSpace(backend)
	}

	// Create config
	cfg := &config.Config{
		Port:     *port,
		Backends: backendList,
	}

	// Create load balancer
	lb, err := balancer.NewLoadBalancer(cfg)
	if err != nil {
		log.Fatalf("Failed to create load balancer: %v", err)
	}

	// Create HTTP server for the load balancer
	loadBalancerServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: lb,
	}

	log.Printf("Load balancer starting on port %d", cfg.Port)
	log.Printf("Forwarding requests to backends: %v", cfg.Backends)

	// Start the load balancer server
	if err := loadBalancerServer.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}
}
