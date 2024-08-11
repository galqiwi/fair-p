package main

import (
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
