package utils

import (
	"io"
	"sync"
)

type Counter struct {
	value int64
	mutex sync.RWMutex

	nSubscribed int64
	subscribers map[int64]func(int64)
}

func NewCounter() *Counter {
	return &Counter{
		subscribers: make(map[int64]func(int64)),
	}
}

func (c *Counter) Add(value int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value += value

	for _, sub := range c.subscribers {
		sub(c.value)
	}
}

func (c *Counter) Sub(value int64) {
	c.Add(-value)
}

func (c *Counter) Get() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.value
}

type countWriter struct {
	c *Counter
}

func (c *Counter) GetCountingWriter() io.Writer {
	return &countWriter{c}
}

func (w *countWriter) Write(p []byte) (int, error) {
	w.c.Add(int64(len(p)))
	return len(p), nil
}
