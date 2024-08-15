package rate_counter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestRateCountingWriter_Write(t *testing.T) {
	interval := 1 * time.Second
	rateCounter := NewRateCountingWriter(interval)

	data := []byte("test data")
	n, err := rateCounter.Write(data)

	require.NoError(t, err)
	assert.Equal(t, len(data), n)
}

func TestRateCountingWriter_Get(t *testing.T) {
	tick := 50 * time.Millisecond
	rateCounter := NewRateCountingWriter(2 * tick)

	data := []byte("test data")
	n, err := rateCounter.Write(data)
	require.NoError(t, err)
	assert.Equal(t, len(data), n)

	assert.Equal(t, Rate(0), rateCounter.GetRate())

	time.Sleep(3 * tick)

	n, err = rateCounter.Write(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)

	assert.Equal(t, Rate(float64(len(data))/(2*tick).Seconds()), rateCounter.GetRate())

	time.Sleep(3 * tick)

	n, err = rateCounter.Write(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)

	assert.Equal(t, Rate(0), rateCounter.GetRate())

	time.Sleep(6 * tick)

	assert.Equal(t, Rate(0), rateCounter.GetRate())
}

func TestRateCountingWriter_Concurrent(t *testing.T) {
	tick := 50 * time.Millisecond
	rateCounter := NewRateCountingWriter(2 * tick)

	data := []byte("test data")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := rateCounter.Write(data)
			assert.NoError(t, err)
			assert.Equal(t, len(data), n)
		}()
	}
	wg.Wait()

	time.Sleep(3 * tick)

	n, err := rateCounter.Write(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.Equal(t, Rate(float64(len(data))/(2*tick).Seconds())*10, rateCounter.GetRate())
		}()
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := rateCounter.Write(data)
			assert.NoError(t, err)
			assert.Equal(t, len(data), n)
		}()
	}
	wg.Wait()

	time.Sleep(2 * tick)
	assert.Equal(t, Rate(float64(len(data))/(2*tick).Seconds())*10, rateCounter.GetRate())
}
