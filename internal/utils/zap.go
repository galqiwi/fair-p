package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"runtime"
)

func NewLogger(options ...zap.Option) (*zap.Logger, error) {
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			// Keys can be anything except the empty string.
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}
	return config.Build(options...)
}

func LogRuntimeInfo(logger *zap.Logger) {
	// Memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// General system information
	numGoroutines := runtime.NumGoroutine()
	numCPU := runtime.NumCPU()
	gomaxprocs := runtime.GOMAXPROCS(0)
	numCgoCalls := runtime.NumCgoCall()

	// Log various runtime and memory statistics
	logger.Info("Runtime Info",
		zap.Int("NumGoroutines", numGoroutines),
		zap.Int("NumCPU", numCPU),
		zap.Int("GOMAXPROCS", gomaxprocs),
		zap.Int64("NumCgoCalls", numCgoCalls),
		zap.Uint64("AllocatedMemory", memStats.Alloc),
		zap.Uint64("TotalAllocatedMemory", memStats.TotalAlloc),
		zap.Uint64("SysMemory", memStats.Sys),
		zap.Uint64("HeapObjects", memStats.HeapObjects),
	)
}
