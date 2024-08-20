package main

import (
	"github.com/galqiwi/fair-p/internal/utils"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

func (run *Runner) handleTunneling(w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	run.concurrentRequests.Add(1)
	defer run.concurrentRequests.Sub(1)

	remoteHost := utils.TryGettingHostFromRemoteAddr(r.RemoteAddr)

	logger = logger.With(
		zap.String("destination", r.Host),
		zap.String("client_host", remoteHost),
		zap.String("client", r.RemoteAddr),
	)

	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		logger.Info("Error dialing destination", zap.String("err", err.Error()))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		logger.Info("Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		logger.Info("Hijacking error", zap.String("err", err.Error()))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	logger.Info("Tunnel established")

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

		logger.Info("Error during copy (send)", zap.String("err", err.Error()))
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

		logger.Info("Error during copy (recv)", zap.String("err", err.Error()))
	}()
	wg.Wait()

	sent := <-sentChan
	recv := <-recvChan

	closingSide := <-closingSideChan

	logger.Info(
		"Tunnel closed",
		zap.Int64("bytes_sent", sent),
		zap.Int64("bytes_received", recv),
		zap.Any("closing_side", closingSide),
	)
}
