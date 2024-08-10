package limiter

import (
	"context"
	"io"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limit defines the maximum transfer speed of data.
// The Limit is represented as a rate limit per second.
// A zero Limit means no transfers are allowed.
type Limit rate.Limit

// Inf represents an infinite rate limit; it allows all transfers, even if the burst is zero.
const Inf = Limit(rate.Inf)

// Every converts the minimum time interval between transfers to a Limit.
func Every(interval time.Duration) Limit {
	return Limit(rate.Every(interval))
}

// A Limiter controls the speed at which transfers are allowed to happen.
//
// The zero value is a valid Limiter, but it will reject all transfers.
// Use New to create non-zero Limiters.
type Limiter struct {
	mu  sync.Mutex
	lim *rate.Limiter
}

// New creates a new Limiter with the given transfer speed limit (bytes/sec).
func New(limit Limit) *Limiter {
	var b int
	if limit != Inf {
		b = int(limit)
	}
	return &Limiter{lim: rate.NewLimiter(rate.Limit(limit), b)}
}

// Limit returns the current limit.
func (lim *Limiter) Limit() Limit {
	return Limit(lim.lim.Limit())
}

// Burst returns the current burst size.
func (lim *Limiter) Burst() int {
	return lim.lim.Burst()
}

// SetLimit sets a new limit.
func (lim *Limiter) SetLimit(newLimit Limit) {
	lim.lim.SetLimit(rate.Limit(newLimit))
}

// SetBurst sets a new burst size.
func (lim *Limiter) SetBurst(newBurst int) {
	lim.lim.SetBurst(newBurst)
}

// waitN waits for availability of n tokens.
func (lim *Limiter) waitN(ctx context.Context, n int) error {
	return lim.lim.WaitN(ctx, n)
}

// Writer returns a writer with limiting.
func (lim *Limiter) Writer(w io.Writer) io.Writer {
	return lim.WriterWithContext(context.Background(), w)
}

// WriterWithContext returns a writer with limiting and context.
func (lim *Limiter) WriterWithContext(ctx context.Context, w io.Writer) io.Writer {
	if lim.Limit() == Inf {
		return w
	}
	return &writer{lim, w, ctx}
}

// Reader returns a reader with limiting.
func (lim *Limiter) Reader(r io.Reader) io.Reader {
	return lim.ReaderWithContext(context.Background(), r)
}

// ReaderWithContext returns a reader with limiting and context.
func (lim *Limiter) ReaderWithContext(ctx context.Context, r io.Reader) io.Reader {
	if lim.Limit() == Inf {
		return r
	}
	return &reader{lim, r, ctx}
}
