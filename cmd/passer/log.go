package main

import (
	"context"
	"fmt"
	"github.com/galqiwi/fair-p/internal/utils"
	"net/http"
	"runtime"
	"time"

	"go.uber.org/zap"
)

func (run *Runner) runRuntimeLogLoop() {
	for {
		run.logRuntimeInfo()
		time.Sleep(run.runtimeLogInterval)
	}
}

func (run *Runner) logRuntimeInfo() {
	// Memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// General system information
	numGoroutines := runtime.NumGoroutine()
	numCPU := runtime.NumCPU()
	gomaxprocs := runtime.GOMAXPROCS(0)
	numCgoCalls := runtime.NumCgoCall()

	// Log various runtime and memory statistics
	run.logger.Info("Runtime Info",
		zap.Float64("UploadSpeed (MB/s)", float64(run.mainSendRateCounter.GetRate()/1024/1024)),
		zap.Float64("DownloadSpeed (MB/s)", float64(run.mainRecvRateCounter.GetRate()/1024/1024)),
		zap.Float64("GuaranteedThroughput(send) (MB/s)", float64(run.hostSendLimiterStorage.GetGuaranteedThroughput()/1024/1024)),
		zap.Float64("GuaranteedThroughput(recv) (MB/s)", float64(run.hostRecvLimiterStorage.GetGuaranteedThroughput()/1024/1024)),
		zap.Int64("BytesSent", run.mainSendBytesCounter.Get()),
		zap.Int64("BytesReceived", run.mainRecvBytesCounter.Get()),
		zap.Int64("ConcurrentRemotes(send)", run.hostSendLimiterStorage.GetNHosts()),
		zap.Int64("ConcurrentRemotes(recv)", run.hostRecvLimiterStorage.GetNHosts()),
		zap.Int64("NumConcurrentRequests", run.concurrentRequests.Get()),
		zap.Int("NumGoroutines", numGoroutines),
		zap.Int64("MainRecvLimiterTokens", int64(run.mainRecvLimiter.Tokens())),
		zap.Int64("MainSendLimiterTokens", int64(run.mainSendLimiter.Tokens())),
		zap.Int("NumCPU", numCPU),
		zap.Int("GOMAXPROCS", gomaxprocs),
		zap.Int64("NumCgoCalls", numCgoCalls),
		zap.Uint64("AllocatedMemory", memStats.Alloc),
		zap.Uint64("TotalAllocatedMemory", memStats.TotalAlloc),
		zap.Uint64("SysMemory", memStats.Sys),
		zap.Uint64("HeapObjects", memStats.HeapObjects),
	)
}

func (run *Runner) logRuntimeInfoHandler(w http.ResponseWriter, r *http.Request) {
	remoteHost := utils.TryGettingHostFromRemoteAddr(r.RemoteAddr)
	hostLimiter := run.hostHealthLimiterStorage.GetLimiterHandle(remoteHost)

	defer func() {
		go func() {
			limiterPeriod := time.Duration(float64(time.Second) / float64(hostLimiter.Limit()))
			time.Sleep(limiterPeriod * 2)
			hostLimiter.CloseHandle()
		}()
	}()

	err := hostLimiter.WaitN(context.Background(), 1)
	if err != nil {
		panic(err)
	}

	// Memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// General system information
	numGoroutines := runtime.NumGoroutine()
	numCPU := runtime.NumCPU()
	gomaxprocs := runtime.GOMAXPROCS(0)
	numCgoCalls := runtime.NumCgoCall()

	_, _ = fmt.Fprintf(w, "UploadSpeed: %.2f MB/s\n", float64(run.mainSendRateCounter.GetRate()/1024/1024))
	_, _ = fmt.Fprintf(w, "DownloadSpeed: %.2f MB/s\n", float64(run.mainRecvRateCounter.GetRate()/1024/1024))
	_, _ = fmt.Fprintf(w, "GuaranteedThroughput(send): %.2f MB/s\n", float64(run.hostSendLimiterStorage.GetGuaranteedThroughput()/1024/1024))
	_, _ = fmt.Fprintf(w, "GuaranteedThroughput(recv): %.2f MB/s\n", float64(run.hostRecvLimiterStorage.GetGuaranteedThroughput()/1024/1024))
	_, _ = fmt.Fprintf(w, "BytesSent: %d\n", run.mainSendBytesCounter.Get())
	_, _ = fmt.Fprintf(w, "BytesReceived: %d\n", run.mainRecvBytesCounter.Get())
	_, _ = fmt.Fprintf(w, "ConcurrentRemotes(send): %d\n", run.hostSendLimiterStorage.GetNHosts())
	_, _ = fmt.Fprintf(w, "ConcurrentRemotes(recv): %d\n", run.hostRecvLimiterStorage.GetNHosts())
	_, _ = fmt.Fprintf(w, "NumConcurrentRequests: %d\n", run.concurrentRequests.Get())
	_, _ = fmt.Fprintf(w, "NumGoroutines: %d\n", numGoroutines)
	_, _ = fmt.Fprintf(w, "MainRecvLimiterTokens: %d\n", int64(run.mainRecvLimiter.Tokens()))
	_, _ = fmt.Fprintf(w, "MainSendLimiterTokens: %d\n", int64(run.mainSendLimiter.Tokens()))
	_, _ = fmt.Fprintf(w, "NumCPU: %d\n", numCPU)
	_, _ = fmt.Fprintf(w, "GOMAXPROCS: %d\n", gomaxprocs)
	_, _ = fmt.Fprintf(w, "NumCgoCalls: %d\n", numCgoCalls)
	_, _ = fmt.Fprintf(w, "AllocatedMemory: %d bytes\n", memStats.Alloc)
	_, _ = fmt.Fprintf(w, "TotalAllocatedMemory: %d bytes\n", memStats.TotalAlloc)
	_, _ = fmt.Fprintf(w, "SysMemory: %d bytes\n", memStats.Sys)
	_, _ = fmt.Fprintf(w, "HeapObjects: %d\n", memStats.HeapObjects)
}
