package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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
