package ratelimit

import (
	"context"
	"io"

	"github.com/neilotoole/fifomu"
	"golang.org/x/time/rate"
)

type reader struct {
	inner   io.Reader
	limiter *rate.Limiter
}

func NewRateLimitedReader(r io.Reader, limiter *rate.Limiter) io.Reader {
	return &reader{
		inner:   r,
		limiter: limiter,
	}
}

var mu fifomu.Mutex

func (r *reader) Read(p []byte) (n int, err error) {
	toRead := len(p)

	burst := r.limiter.Burst()

	if toRead > burst {
		toRead = burst
	}

	if toRead <= 0 {
		panic("invalid Read")
	}

	n, err = r.inner.Read(p[:toRead])

	mu.Lock()
	waitErr := r.limiter.WaitN(context.Background(), n)
	mu.Unlock()

	if waitErr != nil {
		panic("invalid limiter.WaitN call")
	}

	return n, err
}
