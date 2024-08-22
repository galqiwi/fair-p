package logutils

import (
	"go.uber.org/zap/zapcore"
	"io"
)

type asyncWriteSyncer struct {
	inner        zapcore.WriteSyncer
	messageQueue chan []byte
}

func NewAsyncWriter(ws zapcore.WriteSyncer, size int) (io.Writer, func() int) {
	output := &asyncWriteSyncer{
		inner:        ws,
		messageQueue: make(chan []byte, size),
	}
	go output.writeLoop()
	return output, func() int {
		return len(output.messageQueue)
	}
}

func (s *asyncWriteSyncer) Write(bs []byte) (int, error) {
	n := len(bs)

	s.messageQueue <- bs

	return n, nil
}

func (s *asyncWriteSyncer) writeLoop() {
	for bs := range s.messageQueue {
		_, _ = s.inner.Write(bs)
	}
}
