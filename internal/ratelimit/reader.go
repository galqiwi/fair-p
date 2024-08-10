package ratelimit

import (
	"context"
	"errors"
	"io"

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

func (r *reader) Read(p []byte) (n int, err error) {
	n, err = r.inner.Read(p)

	waitErr := r.limiter.WaitN(context.Background(), n)

	err = errors.Join(err, waitErr)

	return n, err
}
