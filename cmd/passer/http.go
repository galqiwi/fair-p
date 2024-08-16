package main

import (
	"github.com/galqiwi/fair-p/internal/ratelimit"
	"github.com/galqiwi/fair-p/internal/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func (run *Runner) handleHTTP(w http.ResponseWriter, r *http.Request, traceId uuid.UUID) {
	run.concurrentRequests.Add(1)
	defer run.concurrentRequests.Sub(1)

	remoteHost := utils.TryGettingHostFromRemoteAddr(r.RemoteAddr)

	run.logger.Info("Handling HTTP request",
		zap.String("url", r.URL.String()),
		zap.String("remote_host", remoteHost),
		zap.String("trace_id", traceId.String()),
	)

	// TODO: upload limiter?
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		run.logger.Info("RoundTrip error",
			zap.String("url", r.URL.String()),
			zap.String("err", err.Error()),
			zap.String("trace_id", traceId.String()),
		)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	utils.CopyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	hostLimiter := run.hostRecvLimiterStorage.GetLimiterHandle(remoteHost)
	defer hostLimiter.CloseHandle()
	recv, err := ratelimit.Copy(
		io.MultiWriter(w, run.mainRecvRateCounter, run.mainRecvBitsCounter.GetCountingWriter()),
		resp.Body,
		[]ratelimit.Limiter{
			ratelimit.NewCombinedLimiter(hostLimiter.Limiter, run.sharedRecvLimiter),
			run.mainRecvLimiter,
		},
	)

	if err != nil {
		run.logger.Info("Error copying response body",
			zap.String("url", r.URL.String()),
			zap.String("err", err.Error()),
			zap.String("trace_id", traceId.String()),
		)
		return
	}
	run.logger.Info("HTTP response forwarded",
		zap.String("remote_host", remoteHost),
		zap.String("url", r.URL.String()),
		zap.Int64("bits_received", recv),
		zap.String("trace_id", traceId.String()),
	)
}
