package ratelimit

import (
	"context"
	"time"
)

type Limiter interface {
	Burst() int
	WaitN(ctx context.Context, n int) (err error)
	AllowN(t time.Time, n int) bool
	Tokens() float64
}

type combinedLimiter struct {
	guaranteed Limiter
	shared     Limiter
}

func NewCombinedLimiter(guaranteed, shared Limiter) Limiter {
	return &combinedLimiter{guaranteed: guaranteed, shared: shared}
}

func (p *combinedLimiter) Burst() int {
	return p.guaranteed.Burst()
}

func (p *combinedLimiter) WaitN(ctx context.Context, n int) (err error) {
	if p.guaranteed.AllowN(time.Now(), n) {
		return nil
	}
	if p.shared.Burst() >= n && p.shared.AllowN(time.Now(), n) {
		return nil
	}
	return p.guaranteed.WaitN(ctx, n)
}

func (p *combinedLimiter) Tokens() float64 {
	return p.guaranteed.Tokens()
}

func (p *combinedLimiter) AllowN(t time.Time, n int) bool {
	panic("not implemented")
}
