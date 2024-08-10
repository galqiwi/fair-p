package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounter_Add(t *testing.T) {
	c := NewCounter()
	c.Add(5)
	assert.Equal(t, int64(5), c.Get())

	c.Add(3)
	assert.Equal(t, int64(8), c.Get())
}

func TestCounter_Sub(t *testing.T) {
	c := NewCounter()
	c.Add(10)
	assert.Equal(t, int64(10), c.Get())

	c.Sub(3)
	assert.Equal(t, int64(7), c.Get())
}

func TestCounter_Get(t *testing.T) {
	c := NewCounter()
	assert.Equal(t, int64(0), c.Get())

	c.Add(10)
	assert.Equal(t, int64(10), c.Get())
}

func TestCounter_Subscribe(t *testing.T) {
	c := NewCounter()
	mutex := sync.Mutex{} // Protects the subscribers map from concurrent updates

	var subscriberCalled bool
	ticket := c.Subscribe(func(value int64) {
		mutex.Lock()
		subscriberCalled = true
		mutex.Unlock()
	})
	require.NotNil(t, ticket)

	c.Add(5)

	mutex.Lock()
	assert.True(t, subscriberCalled)
	mutex.Unlock()
}

func TestTicket_Unsubscribe(t *testing.T) {
	c := NewCounter()
	mutex := sync.Mutex{} // Protects the subscribers map from concurrent updates

	var subscriberCalled bool
	ticket := c.Subscribe(func(value int64) {
		mutex.Lock()
		subscriberCalled = true
		mutex.Unlock()
	})
	require.NotNil(t, ticket)

	ticket.Unsubscribe()
	c.Add(5)

	mutex.Lock()
	assert.False(t, subscriberCalled)
	mutex.Unlock()
}

// Concurrent tests
func TestCounter_ConcurrentAdd(t *testing.T) {
	c := NewCounter()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Add(1)
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(10), c.Get())
}

func TestCounter_ConcurrentSub(t *testing.T) {
	c := NewCounter()
	c.Add(10)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Sub(1)
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(0), c.Get())
}

func TestCounter_ConcurrentSubscribe(t *testing.T) {
	c := NewCounter()
	var wg sync.WaitGroup
	mutex := sync.Mutex{}
	subscriberCount := 0

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Subscribe(func(value int64) {
				mutex.Lock()
				subscriberCount++
				mutex.Unlock()
			})
		}()
	}

	wg.Wait()
	assert.Equal(t, 10, len(c.subscribers))

	assert.Equal(t, 0, subscriberCount)
	c.Add(1)
	assert.Equal(t, 10, subscriberCount)
}
