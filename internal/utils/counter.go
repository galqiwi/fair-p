package utils

import "sync"

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

type Ticket struct {
	counter *Counter
	number  int64
}

func (c *Counter) Subscribe(callback func(value int64)) *Ticket {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	tickerNumber := c.nSubscribed
	c.nSubscribed += 1

	c.subscribers[tickerNumber] = callback

	return &Ticket{counter: c, number: tickerNumber}
}

func (t *Ticket) Unsubscribe() {
	c := t.counter

	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.subscribers, t.number)
}
