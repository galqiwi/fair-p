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
	noIPv4             bool

	concurrentRequests       *utils.Counter
	hostHealthLimiterStorage *hostlimiters.HostLimiterStorage
	hostSendLimiterStorage   *hostlimiters.HostLimiterStorage
	hostRecvLimiterStorage   *hostlimiters.HostLimiterStorage
	logger                   *zap.Logger
	mainSendLimiter          ratelimit.Limiter
	sharedSendLimiter        ratelimit.Limiter
	mainSendRateCounter      *rate_counter.RateCountingWriter
	mainSendBytesCounter     *utils.Counter
	mainRecvLimiter          ratelimit.Limiter
	sharedRecvLimiter        ratelimit.Limiter
	mainRecvRateCounter      *rate_counter.RateCountingWriter
	mainRecvBytesCounter     *utils.Counter
}

func NewRunner(a args) (*Runner, error) {
	burstSize := 2 * 1024 * 1024
	rateCounterDuration := time.Second

	healthLimit := rate.Every(time.Second)
	healthBurst := 3

	logger, err := utils.NewLogger()
	if err != nil {
		return nil, err
	}
	return &Runner{
		runtimeLogInterval: a.runtimeLogInterval,
		port:               a.port,
		noIPv4:             a.noIPv4,

		concurrentRequests:       utils.NewCounter(),
		hostHealthLimiterStorage: hostlimiters.NewHostLimiterStorage(healthLimit, healthBurst),
		hostSendLimiterStorage:   hostlimiters.NewHostLimiterStorage(a.maxThroughput/2, burstSize),
		hostRecvLimiterStorage:   hostlimiters.NewHostLimiterStorage(a.maxThroughput/2, burstSize),
		logger:                   logger,
		mainSendLimiter:          rate.NewLimiter(a.maxThroughput, burstSize),
		sharedSendLimiter:        rate.NewLimiter(a.maxThroughput/2, burstSize),
		mainSendRateCounter:      rate_counter.NewRateCountingWriter(rateCounterDuration),
		mainSendBytesCounter:     utils.NewCounter(),
		mainRecvLimiter:          rate.NewLimiter(a.maxThroughput, burstSize),
		sharedRecvLimiter:        rate.NewLimiter(a.maxThroughput/2, burstSize),
		mainRecvRateCounter:      rate_counter.NewRateCountingWriter(rateCounterDuration),
		mainRecvBytesCounter:     utils.NewCounter(),
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

	logger := run.logger.With(zap.String("trace_id", traceId.String()))

	utils.LogHttpRequest(logger, r)

	if r.Method == http.MethodConnect {
		run.handleTunneling(w, r, logger)
		return
	}

	if strings.HasPrefix(r.URL.String(), "/register") {
		logger.Info(
			"Registered host",
			zap.String("url", r.URL.String()),
			zap.String("client", r.RemoteAddr),
		)
		_, _ = fmt.Fprintf(w, "Thank you for registering :)\n")
		return
	}

	if strings.HasPrefix(r.URL.String(), "/health") {
		run.logRuntimeInfoHandler(w, r)
		return
	}

	run.handleHTTP(w, r, logger)
}
