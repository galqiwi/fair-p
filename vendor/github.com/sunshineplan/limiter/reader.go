package limiter

import (
	"context"
	"io"
)

var _ io.Reader = &reader{}

type reader struct {
	*Limiter
	r   io.Reader
	ctx context.Context
}

func (r *reader) Read(p []byte) (int, error) {
	if r.Limit() == Inf {
		return r.r.Read(p)
	}
	burst := r.Burst()
	if len(p) <= burst {
		n, err := r.r.Read(p)
		if err := r.waitN(r.ctx, n); err != nil {
			return n, err
		}
		return n, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	var read int
	for i := 0; i < len(p); i += burst {
		end := i + burst
		if end > len(p) {
			end = len(p)
		}
		n, err := r.r.Read(p[i:end])
		read += n
		if err := r.waitN(r.ctx, n); err != nil {
			return read, err
		}
		if err != nil {
			return read, err
		}
	}
	return read, nil
}
