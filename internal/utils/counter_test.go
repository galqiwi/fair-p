package utils

import (
	"sync"
	"testing"
)

func TestCounter_Add(t *testing.T) {
	t.Run("test adding positive value", func(t *testing.T) {
		c := Counter{}

		c.Add(10)
		if got := c.Get(); got != 10 {
			t.Errorf("Counter.Add() = %v, want %v", got, 10)
		}
	})

	t.Run("test adding negative value", func(t *testing.T) {
		c := Counter{}

		c.Add(-5)
		if got := c.Get(); got != -5 {
			t.Errorf("Counter.Add() = %v, want %v", got, -5)
		}
	})
}

func TestCounter_Sub(t *testing.T) {
	t.Run("test subtracting positive value", func(t *testing.T) {
		c := Counter{}
		c.Add(10)

		c.Sub(5)
		if got := c.Get(); got != 5 {
			t.Errorf("Counter.Sub() = %v, want %v", got, 5)
		}
	})

	t.Run("test subtracting negative value", func(t *testing.T) {
		c := Counter{}
		c.Add(10)

		c.Sub(-5)
		if got := c.Get(); got != 15 {
			t.Errorf("Counter.Sub() = %v, want %v", got, 15)
		}
	})
}

func TestCounter_Get(t *testing.T) {
	t.Run("test get value", func(t *testing.T) {
		c := Counter{}
		c.Add(25)

		if got := c.Get(); got != 25 {
			t.Errorf("Counter.Get() = %v, want %v", got, 25)
		}
	})

	t.Run("test get value with concurrency", func(t *testing.T) {
		c := Counter{}
		var wg sync.WaitGroup

		// Add and Subtract concurrently to test thread safety
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				c.Add(1)
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				c.Sub(1)
			}()
		}
		wg.Wait()

		if got := c.Get(); got != 0 {
			t.Errorf("Counter.Get() with concurrency = %v, want %v", got, 0)
		}
	})
}
