package ratelimit

import (
	"context"
	"io"
)

type reader struct {
	inner   io.Reader
	limiter *FairLimiter
}

func NewRateLimitedReader(r io.Reader, limiter *FairLimiter) io.Reader {
	return &reader{
		inner:   r,
		limiter: limiter,
	}
}

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

	waitErr := r.limiter.WaitN(context.Background(), n)

	if waitErr != nil {
		panic("invalid limiter.WaitN call")
	}

	return n, err
}
