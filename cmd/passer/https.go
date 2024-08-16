package main

import (
	"github.com/galqiwi/fair-p/internal/utils"
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

	remoteHost := utils.TryGettingHostFromRemoteAddr(r.RemoteAddr)

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

	closingSideChan := make(chan string, 2)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			closingSideChan <- "send"
			destConn.Close()
			clientConn.Close()
		}()

		n, err := run.CopySend(destConn, clientConn, remoteHost)

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
		defer func() {
			closingSideChan <- "recv"
			destConn.Close()
			clientConn.Close()
		}()

		n, err := run.CopyRecv(clientConn, destConn, remoteHost)

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

	closingSide := <-closingSideChan

	run.logger.Info(
		"Tunnel closed",
		zap.String("client", r.RemoteAddr),
		zap.String("destination", r.Host),
		zap.Int64("bits_sent", sent),
		zap.Int64("bits_received", recv),
		zap.String("trace_id", traceId.String()),
		zap.String("remote_host", remoteHost),
		zap.Any("closing_side", closingSide),
	)
}
