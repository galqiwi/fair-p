package ratelimit

import "context"

type Limiter interface {
	Burst() int
	WaitN(ctx context.Context, n int) (err error)
}
