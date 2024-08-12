package main

import (
	"fmt"
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
		zap.Float64("GuaranteedThroughput(send) (MB/s)", float64(run.hostSendLimiterStorage.GetGuaranteedThroughput()/1024/1024)),
		zap.Float64("GuaranteedThroughput(recv) (MB/s)", float64(run.hostRecvLimiterStorage.GetGuaranteedThroughput()/1024/1024)),
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
	// Memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// General system information
	numGoroutines := runtime.NumGoroutine()
	numCPU := runtime.NumCPU()
	gomaxprocs := runtime.GOMAXPROCS(0)
	numCgoCalls := runtime.NumCgoCall()

	// Output various runtime and memory statistics to the response writer
	_, _ = fmt.Fprintf(w, "Runtime Info:\n")
	_, _ = fmt.Fprintf(w, "GuaranteedThroughput(send): %.2f MB/s\n", float64(run.hostSendLimiterStorage.GetGuaranteedThroughput()/1024/1024))
	_, _ = fmt.Fprintf(w, "GuaranteedThroughput(recv): %.2f MB/s\n", float64(run.hostRecvLimiterStorage.GetGuaranteedThroughput()/1024/1024))
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
