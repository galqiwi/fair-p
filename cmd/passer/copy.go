package main

import (
	"fmt"
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
		run.getSelfLimit(run.concurrentRemotes.Get()),
		run.burstSize,
	)

	token := run.concurrentRemotes.Subscribe(func(value int64) {
		fmt.Println(run.getSelfLimit(value))
		selfLimiter.SetLimit(run.getSelfLimit(value))
	})
	defer token.Unsubscribe()

	n, err := ratelimit.Copy(dst, src, []*rate.Limiter{selfLimiter, mainLimiter})

	return n, err
}

func (run *Runner) getSelfLimit(nConcurrentRemotes int64) rate.Limit {
	if nConcurrentRemotes <= 0 {
		panic(fmt.Sprintf("invalid number of concurrent requests: %d", nConcurrentRemotes))
	}
	return rate.Limit(float64(run.maxThroughput) / float64(nConcurrentRemotes))
}
