package main

import (
	"github.com/galqiwi/fair-p/internal/ratelimit"
	"github.com/galqiwi/fair-p/internal/utils"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (run *Runner) handleTunneling(w http.ResponseWriter, r *http.Request, traceId uuid.UUID) {
	run.concurrentRequests.Add(1)
	defer run.concurrentRequests.Sub(1)

	remoteHost, err := utils.GetHostFromRemoteAddr(r.RemoteAddr)
	if err != nil {
		run.logger.Info("Failed to parse RemoteAddr",
			zap.String("err", err.Error()),
			zap.String("trace_id", traceId.String()))
		remoteHost = "UNKNOWN_HOST"
	}

	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		run.logger.Info("Error dialing destination", zap.String("host", r.Host), zap.String("err", err.Error()),
			zap.String("trace_id", traceId.String()), zap.String("remote_host", remoteHost))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		run.logger.Info("Hijacking not supported",
			zap.String("trace_id", traceId.String()), zap.String("remote_host", remoteHost))
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		run.logger.Info("Hijacking error", zap.String("err", err.Error()),
			zap.String("trace_id", traceId.String()), zap.String("remote_host", remoteHost))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	run.logger.Info("Tunnel established", zap.String("client", r.RemoteAddr), zap.String("destination", r.Host),
		zap.String("trace_id", traceId.String()), zap.String("remote_host", remoteHost))

	sentChan := make(chan int64, 1)
	recvChan := make(chan int64, 1)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer destConn.Close()
		defer clientConn.Close()

		hostLimiter := run.hostSendLimiterStorage.GetLimiterHandle(remoteHost)
		defer hostLimiter.CloseHandle()

		n, err := ratelimit.Copy(
			io.MultiWriter(destConn, run.mainSendCounter, run.mainSendBitsCounter.GetCountingWriter()),
			clientConn,
			[]ratelimit.Limiter{
				ratelimit.NewCombinedLimiter(hostLimiter.Limiter, run.sharedSendLimiter),
				run.mainSendLimiter,
			},
		)

		sentChan <- n

		if err == nil {
			return
		}

		run.logger.Info(
			"Error during copy (send)",
			zap.String("trace_id", traceId.String()),
			zap.String("err", err.Error()),
		)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer destConn.Close()
		defer clientConn.Close()

		hostLimiter := run.hostRecvLimiterStorage.GetLimiterHandle(remoteHost)
		defer hostLimiter.CloseHandle()

		n, err := ratelimit.Copy(
			io.MultiWriter(clientConn, run.mainRecvCounter, run.mainRecvBitsCounter.GetCountingWriter()),
			destConn,
			[]ratelimit.Limiter{
				ratelimit.NewCombinedLimiter(hostLimiter.Limiter, run.sharedRecvLimiter),
				run.mainRecvLimiter,
			},
		)

		recvChan <- n

		if err == nil {
			return
		}

		run.logger.Info(
			"Error during copy (recv)",
			zap.String("trace_id", traceId.String()),
			zap.String("err", err.Error()),
		)
	}()
	wg.Wait()

	sent := <-sentChan
	recv := <-recvChan

	run.logger.Info(
		"Tunnel closed",
		zap.String("client", r.RemoteAddr),
		zap.String("destination", r.Host),
		zap.Int64("bits_sent", sent),
		zap.Int64("bits_received", recv),
		zap.String("trace_id", traceId.String()),
		zap.String("remote_host", remoteHost),
	)
}
