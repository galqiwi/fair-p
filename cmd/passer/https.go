package main

import (
	"bufio"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

func (run *Runner) handleTunneling(w http.ResponseWriter, r *http.Request, traceId uuid.UUID) {
	run.concurrentRequests.Add(1)
	defer run.concurrentRequests.Sub(1)

	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		run.logger.Info("Error dialing destination", zap.String("host", r.Host), zap.String("err", err.Error()),
			zap.String("trace_id", traceId.String()))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		run.logger.Info("Hijacking not supported",
			zap.String("trace_id", traceId.String()))
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		run.logger.Info("Hijacking error", zap.String("err", err.Error()),
			zap.String("trace_id", traceId.String()))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	run.logger.Info("Tunnel established", zap.String("client", r.RemoteAddr), zap.String("destination", r.Host),
		zap.String("trace_id", traceId.String()))

	var sent int64 = 0
	var recv int64 = 0

	defer destConn.Close()
	defer clientConn.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		clientConn := bufio.NewReader(clientConn)
		destConn := bufio.NewWriter(destConn)

		var err error
		sent, err = io.Copy(destConn, clientConn)
		if err == nil {
			return
		}

		run.logger.Info(
			"Error during io.Copy(destConn, clientConn)",
			zap.String("client", r.RemoteAddr),
			zap.String("destination", r.Host),
			zap.String("trace_id", traceId.String()),
			zap.String("err", err.Error()),
		)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()

		destConn := bufio.NewReader(destConn)
		clientConn := bufio.NewWriter(clientConn)

		var err error
		recv, err = io.Copy(clientConn, destConn)
		if err == nil {
			return
		}

		run.logger.Info(
			"Error during io.Copy(clientConn, destConn)",
			zap.String("client", r.RemoteAddr),
			zap.String("destination", r.Host),
			zap.String("trace_id", traceId.String()),
			zap.String("err", err.Error()),
		)
	}()
	wg.Wait()

	run.logger.Info(
		"Tunnel closed",
		zap.String("client", r.RemoteAddr),
		zap.String("destination", r.Host),
		zap.Int64("bits_sent", sent),
		zap.Int64("bits_received", recv),
		zap.String("trace_id", traceId.String()),
	)
}
