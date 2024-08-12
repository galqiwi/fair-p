package rate_counter

import (
	"io"
	"sync"
	"time"
)

type Rate float64

// RateCountingWriter is an io.Writer that counts the number of bits written over defined intervals.
// This allows for tracking the writing rate (bits per second) over time.
type RateCountingWriter struct {
	mu sync.RWMutex

	intervalDuration time.Duration

	lastIntervalFinish time.Time
	lastIntervalBits   int64

	thisIntervalBits int64
}

var _ io.Writer = (*RateCountingWriter)(nil)

func NewRateCountingWriter(intervalDuration time.Duration) *RateCountingWriter {
	return &RateCountingWriter{
		intervalDuration: intervalDuration,

		lastIntervalFinish: time.Now(),
		lastIntervalBits:   0,

		thisIntervalBits: 0,
	}
}

func (r *RateCountingWriter) Write(p []byte) (n int, err error) {
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	if now.After(r.lastIntervalFinish.Add(r.intervalDuration)) {
		r.lastIntervalFinish = now
		r.lastIntervalBits = r.thisIntervalBits
		r.thisIntervalBits = 0
	}

	if now.After(r.lastIntervalFinish) {
		r.thisIntervalBits += int64(len(p))
	} else {
		r.lastIntervalBits += int64(len(p))
	}

	return len(p), nil
}

func (r *RateCountingWriter) GetRate() Rate {
	now := time.Now()

	r.mu.RLock()
	defer r.mu.RUnlock()

	for now.After(r.lastIntervalFinish.Add(r.intervalDuration * 2)) {
		r.mu.RUnlock()
		// updates interval statistics
		_, _ = r.Write(nil)
		now = time.Now()
		r.mu.RLock()
	}

	return Rate(float64(r.lastIntervalBits) / r.intervalDuration.Seconds())
}
