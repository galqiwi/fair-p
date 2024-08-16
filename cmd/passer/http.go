package main

import (
	"github.com/galqiwi/fair-p/internal/utils"
	"go.uber.org/zap"
	"net/http"
)

func (run *Runner) handleHTTP(w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	run.concurrentRequests.Add(1)
	defer run.concurrentRequests.Sub(1)

	remoteHost := utils.TryGettingHostFromRemoteAddr(r.RemoteAddr)

	logger = logger.With(
		zap.String("url", r.URL.String()),
		zap.String("destination", r.Host),
		zap.String("client_host", remoteHost),
		zap.String("client", r.RemoteAddr),
	)

	logger.Info("Handling HTTP request")

	// TODO: upload limiter?
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		logger.Info("RoundTrip error", zap.String("err", err.Error()))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	utils.CopyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	recv, err := run.CopyRecv(w, resp.Body, remoteHost)

	if err != nil {
		logger.Info("Error copying response body", zap.String("err", err.Error()))
		return
	}
	logger.Info("HTTP response forwarded",
		zap.Int64("bits_received", recv),
	)
}
