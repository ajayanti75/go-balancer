package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// Config holds the configuration for our load balancer
type Config struct {
	Port    int
	Backend string
}

// LoadBalancer represents our basic load balancer
type LoadBalancer struct {
	config     *Config
	client     *http.Client
	backendURL *url.URL
}

// NewLoadBalancer creates a new LoadBalancer instance
func NewLoadBalancer(config *Config) (*LoadBalancer, error) {
	backendURL, err := url.Parse(config.Backend)
	if err != nil {
		return nil, fmt.Errorf("invalid backend URL: %w", err)
	}

	return &LoadBalancer{
		config:     config,
		client:     &http.Client{},
		backendURL: backendURL,
	}, nil
}

// ServeHTTP implements the http.Handler interface
// This is where we handle incoming requests and forward them to the backend
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request on %s from %s:",
		r.Method, r.URL.Path, r.RemoteAddr)
	log.Printf("Host: %s", r.Host)
	log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))

	// Create a new request to forward to the backend
	backendReq, err := http.NewRequest(r.Method, lb.backendURL.String()+r.URL.Path, r.Body)
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
		log.Printf("Error forwarding request to backend: %v", err)
		http.Error(w, "Backend Server Error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Log the response from backend
	log.Printf("Response from server: %s", resp.Status)

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

func main() {
	// Parse command line flags
	var (
		port    = flag.Int("port", 8000, "Port to listen on")
		backend = flag.String("backend", "http://localhost:8080", "Backend server URL")
	)
	flag.Parse()

	// Create config
	config := &Config{
		Port:    *port,
		Backend: *backend,
	}

	// Create load balancer
	lb, err := NewLoadBalancer(config)
	if err != nil {
		log.Fatalf("Failed to create load balancer: %v", err)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: lb,
	}

	log.Printf("Load balancer starting on port %d", config.Port)
	log.Printf("Forwarding requests to: %s", config.Backend)

	// Start the server
	// The server automatically handles concurrency via goroutines for each request
	if err := server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}
}
