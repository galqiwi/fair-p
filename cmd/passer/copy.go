package main

import (
	"github.com/galqiwi/fair-p/internal/ratelimit"
	"io"
)

func (run *Runner) CopyRecv(dst io.Writer, src io.Reader, remoteHost string) (int64, error) {
	hostLimiter := run.hostRecvLimiterStorage.GetLimiterHandle(remoteHost)
	defer hostLimiter.CloseHandle()
	return ratelimit.Copy(
		io.MultiWriter(dst, run.mainRecvRateCounter, run.mainRecvBitsCounter.GetCountingWriter()),
		src,
		[]ratelimit.Limiter{
			ratelimit.NewCombinedLimiter(hostLimiter.Limiter, run.sharedRecvLimiter),
			run.mainRecvLimiter,
		},
	)
}

func (run *Runner) CopySend(dst io.Writer, src io.Reader, remoteHost string) (int64, error) {
	hostLimiter := run.hostSendLimiterStorage.GetLimiterHandle(remoteHost)
	defer hostLimiter.CloseHandle()
	return ratelimit.Copy(
		io.MultiWriter(dst, run.mainSendRateCounter, run.mainSendBitsCounter.GetCountingWriter()),
		src,
		[]ratelimit.Limiter{
			ratelimit.NewCombinedLimiter(hostLimiter.Limiter, run.sharedSendLimiter),
			run.mainSendLimiter,
		},
	)
}
