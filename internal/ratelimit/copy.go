package ratelimit

import (
	"io"
)

func Copy(dst io.Writer, src io.Reader, limiters []Limiter) (written int64, err error) {
	for _, limiter := range limiters {
		src = NewRateLimitedReader(src, limiter)
	}

	return io.Copy(dst, src)
}
