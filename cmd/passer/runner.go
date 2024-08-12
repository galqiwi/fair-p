package main

import (
	"crypto/tls"
	"fmt"
	"github.com/galqiwi/fair-p/internal/hostlimiters"
	"github.com/galqiwi/fair-p/internal/rate_counter"
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
	sharedSendLimiter      ratelimit.Limiter
	mainSendCounter        *rate_counter.RateCountingWriter
	mainRecvLimiter        ratelimit.Limiter
	sharedRecvLimiter      ratelimit.Limiter
	mainRecvCounter        *rate_counter.RateCountingWriter
}

func NewRunner(a args) (*Runner, error) {
	burstSize := 2 * 1024 * 1024
	rateCounterDuration := time.Second

	logger, err := utils.NewLogger()
	if err != nil {
		return nil, err
	}
	return &Runner{
		runtimeLogInterval: a.runtimeLogInterval,
		port:               a.port,

		concurrentRequests:     utils.NewCounter(),
		hostSendLimiterStorage: hostlimiters.NewHostLimiterStorage(a.maxThroughput/2, burstSize),
		hostRecvLimiterStorage: hostlimiters.NewHostLimiterStorage(a.maxThroughput/2, burstSize),
		logger:                 logger,
		mainSendLimiter:        rate.NewLimiter(a.maxThroughput, burstSize),
		sharedSendLimiter:      rate.NewLimiter(a.maxThroughput/2, burstSize),
		mainSendCounter:        rate_counter.NewRateCountingWriter(rateCounterDuration),
		mainRecvLimiter:        rate.NewLimiter(a.maxThroughput, burstSize),
		sharedRecvLimiter:      rate.NewLimiter(a.maxThroughput/2, burstSize),
		mainRecvCounter:        rate_counter.NewRateCountingWriter(rateCounterDuration),
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

	if strings.HasPrefix(r.URL.String(), "/health") {
		run.logRuntimeInfoHandler(w, r)
		return
	}

	run.handleHTTP(w, r, traceId)
}
