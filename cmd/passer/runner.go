package main

import (
	"crypto/tls"
	"fmt"
	"github.com/galqiwi/fair-p/internal/hostlimiters"
	"github.com/galqiwi/fair-p/internal/ratelimit"
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
	port               int

	concurrentRequests     *utils.Counter
	hostSendLimiterStorage *hostlimiters.HostLimiterStorage
	hostRecvLimiterStorage *hostlimiters.HostLimiterStorage
	logger                 *zap.Logger
	mainSendLimiter        ratelimit.Limiter
	mainRecvLimiter        ratelimit.Limiter
}

func NewRunner(a args) (*Runner, error) {
	burstSize := 2 * 1024 * 1024

	logger, err := utils.NewLogger()
	if err != nil {
		return nil, err
	}
	return &Runner{
		runtimeLogInterval: a.runtimeLogInterval,
		port:               a.port,

		concurrentRequests:     utils.NewCounter(),
		hostSendLimiterStorage: hostlimiters.NewHostLimiterStorage(a.maxThroughput, burstSize),
		hostRecvLimiterStorage: hostlimiters.NewHostLimiterStorage(a.maxThroughput, burstSize),
		logger:                 logger,
		mainSendLimiter: ratelimit.NewCombinedLimiter(
			rate.NewLimiter(a.maxThroughput/2, burstSize),
			rate.NewLimiter(a.maxThroughput/2, burstSize)),
		mainRecvLimiter: ratelimit.NewCombinedLimiter(
			rate.NewLimiter(a.maxThroughput/2, burstSize),
			rate.NewLimiter(a.maxThroughput/2, burstSize)),
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
		_, _ = fmt.Fprintf(w, "Thank you for registering :)\n")
		return
	}

	run.handleHTTP(w, r, traceId)
}
