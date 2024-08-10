package internal

import (
	"golang.org/x/time/rate"
	"sync"
)

type HostStorage struct {
	mu       sync.RWMutex
	limiters map[string]rate.Limiter
}
