package limiter

import (
	"context"
	"io"
)

var _ io.Writer = &writer{}

type writer struct {
	*Limiter
	w   io.Writer
	ctx context.Context
}

func (w *writer) Write(p []byte) (int, error) {
	if w.Limit() == Inf {
		return w.w.Write(p)
	}
	burst := w.Burst()
	if len(p) <= burst {
		n, err := w.w.Write(p)
		if err := w.waitN(w.ctx, n); err != nil {
			return n, err
		}
		return n, err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	var written int
	for i := 0; i < len(p); i += burst {
		end := i + burst
		if end > len(p) {
			end = len(p)
		}
		n, err := w.w.Write(p[i:end])
		written += n
		if err := w.waitN(w.ctx, n); err != nil {
			return written, err
		}
		if err != nil {
			return written, err
		}
	}
	return written, nil
}
