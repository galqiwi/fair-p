package utils

import "sync"

type Counter struct {
	value int64
	mutex sync.RWMutex
}

func (c *Counter) Add(value int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value += value
}

func (c *Counter) Sub(value int64) {
	c.Add(-value)
}

func (c *Counter) Get() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.value
}
