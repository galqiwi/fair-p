package utils

import (
	"fmt"
	"sync"
)

// UniqueStringsCounter is a container that stores multiple strings
// (not necessary unique) and can answer in O(1) a number of unique strings inside
type UniqueStringsCounter struct {
	mutex sync.RWMutex

	stringCounter map[string]int64

	nSubscribed int64
	subscribers map[int64]func(int64)
}

func NewUniqueStringsCounter() *UniqueStringsCounter {
	return &UniqueStringsCounter{
		stringCounter: make(map[string]int64),

		subscribers: make(map[int64]func(int64)),
	}
}

func (c *UniqueStringsCounter) Add(value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	oldStringCount := c.stringCounter[value]
	newStringCount := oldStringCount + 1

	c.stringCounter[value] = newStringCount

	if oldStringCount != 0 {
		return
	}

	uniqueCount := int64(len(c.stringCounter))
	for _, sub := range c.subscribers {
		sub(uniqueCount)
	}
}

func (c *UniqueStringsCounter) Remove(value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	oldStringCount := c.stringCounter[value]
	if oldStringCount <= 0 {
		panic(fmt.Sprintf("Cannot remove value %s from UniqueStringsCounter", value))
	}
	newStringCount := oldStringCount - 1

	c.stringCounter[value] = newStringCount

	if newStringCount != 0 {
		return
	}

	delete(c.stringCounter, value)

	uniqueCount := int64(len(c.stringCounter))
	for _, sub := range c.subscribers {
		sub(uniqueCount)
	}
}

func (c *UniqueStringsCounter) Get() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return int64(len(c.stringCounter))
}

func (c *UniqueStringsCounter) Subscribe(callback func(value int64)) *Ticket {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	tickerNumber := c.nSubscribed
	c.nSubscribed += 1

	c.subscribers[tickerNumber] = callback

	return &Ticket{counter: c, number: tickerNumber}
}

type Ticket struct {
	counter *UniqueStringsCounter
	number  int64
}

func (t *Ticket) Unsubscribe() {
	c := t.counter

	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.subscribers, t.number)
}
