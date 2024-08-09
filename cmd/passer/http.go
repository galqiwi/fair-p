package main

import (
	"github.com/galqiwi/fair-p/internal/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func (run *Runner) handleHTTP(w http.ResponseWriter, req *http.Request, traceId uuid.UUID) {
	run.concurrentRequests.Add(1)
	defer run.concurrentRequests.Sub(1)

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		run.logger.Info("RoundTrip error", zap.String("url", req.URL.String()), zap.String("err", err.Error()),
			zap.String("trace_id", traceId.String()))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	utils.CopyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	bytesCopied, err := io.Copy(w, resp.Body)

	if err != nil {
		run.logger.Info("Error copying response body", zap.String("url", req.URL.String()), zap.String("err", err.Error()),
			zap.String("trace_id", traceId.String()))
		return
	}
	run.logger.Info("HTTP response forwarded", zap.String("url", req.URL.String()), zap.Int64("bytes_copied", bytesCopied),
		zap.String("trace_id", traceId.String()))
}
