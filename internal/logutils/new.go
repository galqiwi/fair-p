package logutils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"strings"
	"time"
)

func NewLogger(options ...zap.Option) (*zap.Logger, func() int, error) {
	// Exists infinitely, no ws.Stop() call
	var ws zapcore.WriteSyncer = os.Stdout

	ws = &zapcore.BufferedWriteSyncer{
		WS:            ws,
		Size:          1024 * 1024,
		FlushInterval: time.Second * 10,
	}

	s, getQueueSize := NewAsyncWriter(ws, 1000)
	ws = zapcore.AddSync(s)

	encoderConfig := zapcore.EncoderConfig{
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
	}

	// Build the core with the buffered writer
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		ws,
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)

	core = zapcore.NewSamplerWithOptions(
		core,
		time.Second,
		10,
		10,
	)

	// Build and return the logger
	return zap.New(core, options...), getQueueSize, nil
}

func LogHttpRequest(logger *zap.Logger, r *http.Request) {
	headers := make([]string, 0, len(r.Header))
	for name, values := range r.Header {
		for _, value := range values {
			headers = append(headers, name+": "+value)
		}
	}

	logger.Info("Got request",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("host", r.Host),
		zap.String("client_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.String("headers", strings.Join(headers, ", ")),
	)
}
