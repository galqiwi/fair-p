package ratelimit

import (
	"context"
	"github.com/neilotoole/fifomu"
	"golang.org/x/time/rate"
)

type FairLimiter struct {
	inner *rate.Limiter

	mu fifomu.Mutex
}

func NewFairLimiter(r rate.Limit, b int) *FairLimiter {
	return &FairLimiter{
		inner: rate.NewLimiter(r, b),
	}
}

func (lim *FairLimiter) WaitN(ctx context.Context, n int) (err error) {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	return lim.inner.WaitN(ctx, n)
}

func (lim *FairLimiter) Burst() int {
	return lim.inner.Burst()
}

func (lim *FairLimiter) Tokens() float64 {
	return lim.inner.Tokens()
}
