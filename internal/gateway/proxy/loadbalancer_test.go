package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundRobinBalancer_MaxConcurrencyUnlimited(t *testing.T) {
	lb := NewRoundRobinBalancer()
	backend := Backend{
		ID:             "backend-unlimited",
		URL:            "http://localhost:8001",
		MaxConcurrency: 0, // unlimited
		ModelName:      "gpt-4",
	}
	lb.AddBackend("gpt-4", backend)

	// Acquire concurrency multiple times
	for i := 0; i < 10; i++ {
		ok := lb.AcquireBackend(backend.ID)
		assert.True(t, ok)
	}

	// Verify we can still get the backend
	b, model, ok := lb.Next("gpt-4", "")
	require.True(t, ok)
	assert.Equal(t, backend.ID, b.ID)
	assert.Equal(t, "gpt-4", model)
}

func TestRoundRobinBalancer_MaxConcurrencyLimited(t *testing.T) {
	lb := NewRoundRobinBalancer()
	backend := Backend{
		ID:             "backend-limited",
		URL:            "http://localhost:8002",
		MaxConcurrency: 2,
		ModelName:      "gpt-4",
	}
	lb.AddBackend("gpt-4", backend)

	// Acquire twice
	ok := lb.AcquireBackend(backend.ID)
	assert.True(t, ok)
	ok = lb.AcquireBackend(backend.ID)
	assert.True(t, ok)

	// Third acquire should fail (AcquireBackend enforces MaxConcurrency > 0 check)
	ok = lb.AcquireBackend(backend.ID)
	assert.False(t, ok)

	// Next() should fail because the only backend is at capacity (busy)
	_, _, ok = lb.Next("gpt-4", "")
	assert.False(t, ok)

	// Release one
	lb.ReleaseBackend(backend.ID)

	// Now Next() should succeed
	b, model, ok := lb.Next("gpt-4", "")
	require.True(t, ok)
	assert.Equal(t, backend.ID, b.ID)
	assert.Equal(t, "gpt-4", model)
}

func TestRoundRobinBalancer_AllUnhealthyFallback(t *testing.T) {
	lb := NewRoundRobinBalancer()
	backend1 := Backend{
		ID:             "backend-1",
		URL:            "http://localhost:8001",
		MaxConcurrency: 0,
		ModelName:      "gpt-4",
	}
	backend2 := Backend{
		ID:             "backend-2",
		URL:            "http://localhost:8002",
		MaxConcurrency: 0,
		ModelName:      "gpt-4",
	}
	lb.AddBackend("gpt-4", backend1)
	lb.AddBackend("gpt-4", backend2)

	// Mark both failed (unhealthy)
	lb.MarkFailed(backend1.ID)
	lb.MarkFailed(backend2.ID)

	// Next should still return the first backend as fallback
	b, model, ok := lb.Next("gpt-4", "")
	require.True(t, ok)
	assert.Equal(t, backend1.ID, b.ID)
	assert.Equal(t, "gpt-4", model)
}
