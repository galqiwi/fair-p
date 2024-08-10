package main

import (
	"io"

	"github.com/galqiwi/fair-p/internal/ratelimit"
	"golang.org/x/time/rate"
)

func (run *Runner) CopyWithLimiters(
	dst io.Writer,
	src io.Reader,
	mainLimiter *rate.Limiter,
) (int64, error) {
	selfLimiter := rate.NewLimiter(
		run.getSelfLimit(run.concurrentRequests.Get()),
		run.burstSize,
	)

	token := run.concurrentRequests.Subscribe(func(value int64) {
		selfLimiter.SetLimit(run.getSelfLimit(value))
	})
	defer token.Unsubscribe()

	n, err := ratelimit.Copy(dst, src, []*rate.Limiter{selfLimiter, mainLimiter})

	return n, err
}

func (run *Runner) getSelfLimit(nConcurrentRequests int64) rate.Limit {
	return rate.Limit(float64(run.maxThroughput) / float64(nConcurrentRequests+1))
}
