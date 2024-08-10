package utils

import (
	"go.uber.org/goleak"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestUniqueStringsCounter_Add_CountUnique(t *testing.T) {
	counter := NewUniqueStringsCounter()

	counter.Add("a")
	counter.Add("b")
	counter.Add("a")

	assert.Equal(t, int64(2), counter.Get(), "Expected 2 unique strings")
}

func TestUniqueStringsCounter_Remove(t *testing.T) {
	counter := NewUniqueStringsCounter()

	counter.Add("a")
	counter.Add("b")
	counter.Add("a")
	counter.Remove("a")

	assert.Equal(t, int64(2), counter.Get(), "Expected 2 unique strings after removing one occurrence of 'a'")

	counter.Remove("a")
	assert.Equal(t, int64(1), counter.Get(), "Expected 1 unique string after removing 'a' completely")

	counter.Remove("b")
	assert.Equal(t, int64(0), counter.Get(), "Expected 0 unique strings after removing 'b'")
}

func TestUniqueStringsCounter_Remove_NonExisting(t *testing.T) {
	counter := NewUniqueStringsCounter()

	assert.Panics(t, func() {
		counter.Remove("a")
	}, "Expected panic when trying to remove non-existing value")
}

func TestUniqueStringsCounter_Subscribe(t *testing.T) {
	counter := NewUniqueStringsCounter()

	var countHistory []int64

	ticket := counter.Subscribe(func(count int64) {
		countHistory = append(countHistory, count)
	})

	assert.Equal(t, int64(0), counter.Get())

	counter.Add("a")
	assert.Equal(t, countHistory, []int64{1})
	counter.Add("a")
	assert.Equal(t, countHistory, []int64{1})
	counter.Add("b")
	assert.Equal(t, countHistory, []int64{1, 2})

	var secondCountHistory []int64
	secondTicket := counter.Subscribe(func(count int64) {
		secondCountHistory = append(secondCountHistory, count)
	})
	defer secondTicket.Unsubscribe()

	counter.Remove("a")
	assert.Equal(t, countHistory, []int64{1, 2})
	assert.Equal(t, secondCountHistory, []int64(nil))
	counter.Remove("a")
	assert.Equal(t, countHistory, []int64{1, 2, 1})
	assert.Equal(t, secondCountHistory, []int64{1})
	counter.Remove("b")
	assert.Equal(t, countHistory, []int64{1, 2, 1, 0})
	assert.Equal(t, secondCountHistory, []int64{1, 0})

	ticket.Unsubscribe()
	counter.Add("a")
	assert.Equal(t, countHistory, []int64{1, 2, 1, 0})
	assert.Equal(t, secondCountHistory, []int64{1, 0, 1})
}

func TestUniqueStringsCounter_ConcurrentAddRemove(t *testing.T) {
	counter := NewUniqueStringsCounter()
	var wg sync.WaitGroup
	iterations := 100
	nEnglishLetters := 26

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			counter.Add(string('a' + rune(i%nEnglishLetters))) // Adds 'a' to 'z' repeatedly
		}(i)
	}

	wg.Wait()
	assert.Equal(t, int64(26), counter.Get(), "Expected 26 unique strings with concurrent Adds")

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			counter.Remove(string('a' + rune(i%26))) // Adds 'a' to 'z' repeatedly
		}(i)
	}
	wg.Wait()

	assert.Equal(t, int64(0), counter.Get(), "Expected 0 unique strings after deletion")
}
