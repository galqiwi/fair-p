package ratelimit

import (
	"golang.org/x/time/rate"
	"io"
)

func Copy(dst io.Writer, src io.Reader, limiters []*rate.Limiter) (written int64, err error) {
	for _, limiter := range limiters {
		src = NewRateLimitedReader(src, limiter)
	}

	return io.Copy(dst, src)
}
