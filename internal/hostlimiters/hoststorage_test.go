package hostlimiters

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestHostLimiterStorage_GetLimiterHandle(t *testing.T) {
	maxThroughput := rate.Limit(10)
	burst := 5
	host := "testHost"

	hls := NewHostLimiterStorage(maxThroughput, burst)
	handle := hls.GetLimiterHandle(host)
	require.NotNil(t, handle.Limiter)

	assert.Equal(t, handle.host, host)
	assert.Equal(t, handle.storage, hls)
	assert.Equal(t, handle.Limiter, hls.limiters[host])
	assert.Equal(t, int64(1), hls.limiterUsage[host])

	handle.CloseHandle()

	_, exists := hls.limiters[host]
	assert.False(t, exists)
	_, exists = hls.limiterUsage[host]
	assert.False(t, exists)
}

func TestHostLimiterStorage_Concurrency(t *testing.T) {
	maxThroughput := rate.Limit(10)
	burst := 5
	host := "testHost"

	hls := NewHostLimiterStorage(maxThroughput, burst)
	wg := sync.WaitGroup{}

	const numGoroutines = 10
	wg.Add(numGoroutines)

	// Get limiter handles concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			handle := hls.GetLimiterHandle(fmt.Sprintf("%s_%d", host, i))
			time.Sleep(time.Millisecond * 10)
			handle.CloseHandle()
		}(i)
	}

	wg.Wait()

	assert.Empty(t, hls.limiters)
	assert.Empty(t, hls.limiterUsage)
}
