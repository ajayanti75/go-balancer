package metrics

import (
	"sync"
	"time"
)

// Metrics holds various load balancer metrics
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	totalRequests     int64
	successfulRequests int64
	failedRequests    int64
	
	// Backend metrics
	backendRequests   map[string]int64
	backendFailures   map[string]int64
	
	// Health check metrics
	healthCheckPasses map[string]int64
	healthCheckFails  map[string]int64
	
	// Current state
	healthyBackends   int
	totalBackends     int
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		backendRequests:   make(map[string]int64),
		backendFailures:   make(map[string]int64),
		healthCheckPasses: make(map[string]int64),
		healthCheckFails:  make(map[string]int64),
	}
}

// RecordRequest records a successful request
func (m *Metrics) RecordRequest(backend string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.totalRequests++
	m.successfulRequests++
	m.backendRequests[backend]++
}

// RecordFailure records a failed request
func (m *Metrics) RecordFailure(backend string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.totalRequests++
	m.failedRequests++
	m.backendFailures[backend]++
}

// RecordHealthCheck records a health check result
func (m *Metrics) RecordHealthCheck(backend string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if success {
		m.healthCheckPasses[backend]++
	} else {
		m.healthCheckFails[backend]++
	}
}

// UpdateBackendCount updates the backend count metrics
func (m *Metrics) UpdateBackendCount(healthy, total int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.healthyBackends = healthy
	m.totalBackends = total
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return MetricsSnapshot{
		TotalRequests:      m.totalRequests,
		SuccessfulRequests: m.successfulRequests,
		FailedRequests:     m.failedRequests,
		HealthyBackends:    m.healthyBackends,
		TotalBackends:      m.totalBackends,
		Timestamp:          time.Now(),
	}
}

// MetricsSnapshot represents a point-in-time view of metrics
type MetricsSnapshot struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	HealthyBackends    int
	TotalBackends      int
	Timestamp          time.Time
}

// SuccessRate returns the success rate as a percentage
func (ms *MetricsSnapshot) SuccessRate() float64 {
	if ms.TotalRequests == 0 {
		return 100.0
	}
	return (float64(ms.SuccessfulRequests) / float64(ms.TotalRequests)) * 100.0
}

// HealthyPercentage returns the percentage of healthy backends
func (ms *MetricsSnapshot) HealthyPercentage() float64 {
	if ms.TotalBackends == 0 {
		return 0.0
	}
	return (float64(ms.HealthyBackends) / float64(ms.TotalBackends)) * 100.0
}
