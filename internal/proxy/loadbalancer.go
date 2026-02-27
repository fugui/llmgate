package proxy

import (
	"sync"
	"sync/atomic"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	Next(modelID string) (string, bool)
	MarkFailed(backend string)
	MarkSuccess(backend string)
}

// RoundRobinBalancer 轮询负载均衡器
type RoundRobinBalancer struct {
	mu       sync.RWMutex
	backends map[string][]Backend // modelID -> backends
	counters map[string]*uint32   // modelID -> counter
	healthy  map[string]bool      // backend -> healthy
}

type Backend struct {
	URL    string
	Weight int
}

func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{
		backends: make(map[string][]Backend),
		counters: make(map[string]*uint32),
		healthy:  make(map[string]bool),
	}
}

func (lb *RoundRobinBalancer) AddBackend(modelID string, backend Backend) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.backends[modelID] = append(lb.backends[modelID], backend)
	lb.healthy[backend.URL] = true

	if lb.counters[modelID] == nil {
		var counter uint32
		lb.counters[modelID] = &counter
	}
}

func (lb *RoundRobinBalancer) Next(modelID string) (string, bool) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	backends, exists := lb.backends[modelID]
	if !exists || len(backends) == 0 {
		return "", false
	}

	// 找到健康的后端
	counter := lb.counters[modelID]
	attempts := len(backends)

	for i := 0; i < attempts; i++ {
		idx := atomic.AddUint32(counter, 1) % uint32(len(backends))
		backend := backends[idx]

		if lb.healthy[backend.URL] {
			return backend.URL, true
		}
	}

	// 所有后端都不健康，返回第一个（降级）
	return backends[0].URL, true
}

func (lb *RoundRobinBalancer) MarkFailed(backend string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.healthy[backend] = false
}

func (lb *RoundRobinBalancer) MarkSuccess(backend string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.healthy[backend] = true
}

func (lb *RoundRobinBalancer) GetHealthyBackends(modelID string) []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var healthy []string
	for _, backend := range lb.backends[modelID] {
		if lb.healthy[backend.URL] {
			healthy = append(healthy, backend.URL)
		}
	}
	return healthy
}
