package hostlimiters

import (
	"fmt"
	"golang.org/x/time/rate"
	"sync"
)

type HostLimiterHandle struct {
	*rate.Limiter

	host    string
	storage *HostLimiterStorage
}

type HostLimiterStorage struct {
	mu sync.RWMutex

	maxThroughput rate.Limit
	burst         int

	limiterUsage map[string]int64
	limiters     map[string]*rate.Limiter
}

func NewHostLimiterStorage(maxThroughput rate.Limit, burst int) *HostLimiterStorage {
	return &HostLimiterStorage{
		maxThroughput: maxThroughput,
		burst:         burst,
		limiterUsage:  make(map[string]int64),
		limiters:      make(map[string]*rate.Limiter),
	}
}

func (s *HostLimiterStorage) GetLimiterHandle(host string) HostLimiterHandle {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldLimiterUsage := s.limiterUsage[host]
	newLimiterUsage := oldLimiterUsage + 1

	s.limiterUsage[host] = newLimiterUsage

	if oldLimiterUsage != 0 {
		limiter := s.limiters[host]
		s.validateInnerMaps(host)
		return HostLimiterHandle{limiter, host, s}
	}

	nHosts := int64(len(s.limiters) + 1)

	newThroughput := getLimit(nHosts, s.maxThroughput)

	for _, limiter := range s.limiters {
		limiter.SetLimit(newThroughput)
	}

	output := rate.NewLimiter(newThroughput, s.burst)
	s.limiters[host] = output

	s.validateInnerMaps(host)
	return HostLimiterHandle{output, host, s}
}

func (l *HostLimiterHandle) CloseHandle() {
	s := l.storage
	host := l.host

	s.mu.Lock()
	defer s.mu.Unlock()

	oldLimiterUsage := s.limiterUsage[host]
	newLimiterUsage := oldLimiterUsage - 1

	if newLimiterUsage < 0 {
		panic("invalid limiter usage")
	}

	s.limiterUsage[host] = newLimiterUsage

	if newLimiterUsage != 0 {
		s.validateInnerMaps(host)
		return
	}

	if _, ok := s.limiters[host]; !ok {
		panic("inner maps desynced")
	}

	delete(s.limiters, host)
	delete(s.limiterUsage, host)

	nHosts := int64(len(s.limiters))

	if nHosts == 0 {
		s.validateInnerMaps(host)
		return
	}

	newThroughput := getLimit(nHosts, s.maxThroughput)

	for _, limiter := range s.limiters {
		limiter.SetLimit(newThroughput)
	}

	s.validateInnerMaps(host)
}

func (s *HostLimiterStorage) validateInnerMaps(host string) {
	if s.mu.TryLock() {
		panic("should be called inside the mutex")
	}

	if len(s.limiters) == len(s.limiterUsage) {
		return
	}

	_, limitersOk := s.limiters[host]
	_, limiterUsageOk := s.limiterUsage[host]

	if limiterUsageOk == limitersOk {
		return
	}

	panic("inner maps desynced'")
}

func getLimit(nHosts int64, maxThroughput rate.Limit) rate.Limit {
	if nHosts <= 0 {
		panic(fmt.Sprintf("invalid number of hosts: %d", nHosts))
	}
	return rate.Limit(float64(maxThroughput) / float64(nHosts))
}
