package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/galqiwi/fair-p/internal/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type Runner struct {
	runtimeLogInterval time.Duration
	maxThroughput      rate.Limit
	mainLimiter        *rate.Limiter
	burstSize          int
	logger             *zap.Logger
	port               int

	concurrentRequests *utils.Counter
}

func NewRunner(a args) (*Runner, error) {
	maxThroughput := rate.Limit(80 * 1024 * 1024 * 1024)
	burstSize := 2 * 1024 * 1024

	logger, err := utils.NewLogger()
	if err != nil {
		return nil, err
	}
	return &Runner{
		runtimeLogInterval: a.runtimeLogInterval,
		maxThroughput:      maxThroughput,
		mainLimiter:        rate.NewLimiter(maxThroughput, burstSize),
		burstSize:          burstSize,
		logger:             logger,
		port:               a.port,

		concurrentRequests: utils.NewCounter(),
	}, nil
}

func (run *Runner) Run() error {
	server := http.Server{
		Addr:    fmt.Sprintf(":%v", run.port),
		Handler: http.HandlerFunc(run.mainHandler),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	go run.runRuntimeLogLoop()

	return server.ListenAndServe()
}

func (run *Runner) mainHandler(w http.ResponseWriter, r *http.Request) {
	traceId := uuid.New()

	utils.LogHttpRequest(run.logger, r, traceId)

	if r.Method == http.MethodConnect {
		run.handleTunneling(w, r, traceId)
		return
	}

	if strings.HasPrefix(r.URL.String(), "/register") {
		run.logger.Info(
			"Registered host",
			zap.String("url", r.URL.String()),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("trace_id", traceId.String()),
		)
		_, _ = fmt.Fprintf(w, "Thank you for registering :) (v2)\n")
		return
	}

	run.handleHTTP(w, r, traceId)
}
