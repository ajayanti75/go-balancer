package pool

import (
	"fmt"
	"net/url"
	"strconv"
	"sync"
)

// Backend represents a single backend server
type Backend struct {
	ID      string
	URL     *url.URL
	Healthy bool
	Port    int
}

// ServerPool manages a collection of backend servers
type ServerPool struct {
	backends []*Backend
	mutex    sync.RWMutex // RWMutex allows multiple readers OR one writer
}

// NewServerPool creates a new server pool
func NewServerPool() *ServerPool {
	return &ServerPool{
		backends: make([]*Backend, 0),
	}
}

// AddBackend adds a new backend server to the pool
func (sp *ServerPool) AddBackend(backendURL string) error {
	sp.mutex.Lock()         // Exclusive lock for writing
	defer sp.mutex.Unlock() // Always unlock when function exits

	parsedURL, err := url.Parse(backendURL)
	if err != nil {
		return fmt.Errorf("invalid backend URL %s: %w", backendURL, err)
	}

	backend := &Backend{
		ID:      fmt.Sprintf("backend-%d", len(sp.backends)+1),
		URL:     parsedURL,
		Healthy: true, // Assume healthy initially
		Port:    getPortFromURL(parsedURL),
	}

	sp.backends = append(sp.backends, backend)
	return nil
}

// RemoveBackend removes a backend by ID
func (sp *ServerPool) RemoveBackend(id string) bool {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	for i, backend := range sp.backends {
		if backend.ID == id {
			// Cleaner slice removal using helper
			sp.backends = removeFromSlice(sp.backends, i)
			return true
		}
	}
	return false
}

// GetBackends returns a copy of all backends (for status checking)
func (sp *ServerPool) GetBackends() []*Backend {
	sp.mutex.RLock() // Read lock - allows multiple concurrent readers
	defer sp.mutex.RUnlock()

	// Return a copy to prevent external modification
	backends := make([]*Backend, len(sp.backends))
	copy(backends, sp.backends)
	return backends
}

// GetBackendByIndex returns a backend at specific index (for round-robin)
func (sp *ServerPool) GetBackendByIndex(index int) *Backend {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	if index >= 0 && index < len(sp.backends) {
		return sp.backends[index]
	}
	return nil
}

// GetHealthyBackendCount returns the number of healthy backends
func (sp *ServerPool) GetHealthyBackendCount() int {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	count := 0
	for _, backend := range sp.backends {
		if backend.Healthy {
			count++
		}
	}
	return count
}

// GetBackendCount returns total number of backends
func (sp *ServerPool) GetBackendCount() int {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()
	return len(sp.backends)
}

// SetBackendHealth updates the health status of a backend
func (sp *ServerPool) SetBackendHealth(id string, healthy bool) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	for _, backend := range sp.backends {
		if backend.ID == id {
			backend.Healthy = healthy
			break
		}
	}
}

// Helper function to remove item from slice (cleaner than manual slice manipulation)
func removeFromSlice(slice []*Backend, index int) []*Backend {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// Helper function to extract port from URL using proper string parsing
func getPortFromURL(u *url.URL) int {
	portStr := u.Port()
	if portStr == "" {
		// Default ports for schemes
		if u.Scheme == "https" {
			return 443
		}
		return 80
	}

	// Use Go's built-in string to int conversion
	port, err := strconv.Atoi(portStr)
	if err != nil {
		// If parsing fails, return default
		return 80
	}
	return port
}
