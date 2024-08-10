package ratelimit

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func testCopy(t *testing.T, smallBurst bool) {
	data := "Hello, World!"
	readerData := strings.NewReader(data)

	burst := len(data)
	if smallBurst {
		burst = 1
	}

	limiter := rate.NewLimiter(100, burst)
	ratelimitedReader := NewRateLimitedReader(readerData, limiter)

	dst := &strings.Builder{}
	n, err := io.Copy(dst, ratelimitedReader)

	require.NoError(t, err, "Expected no error or EOF")
	require.Equal(t, len(data), int(n), "Expected to read correct number of bytes")

	output := dst.String()
	assert.Equal(t, data, output, "Expected to read correct data")
}

func TestRateLimitedReader_Read(t *testing.T) {
	t.Run("NormalBurst", func(t *testing.T) {
		testCopy(t, false)
	})

	t.Run("SmallBurst", func(t *testing.T) {
		testCopy(t, true)
	})
}

func TestCopy_NoLimiter(t *testing.T) {
	srcContent := []byte("test content")
	src := bytes.NewReader(srcContent)
	dst := &bytes.Buffer{}

	written, err := Copy(dst, src, []*rate.Limiter{})
	require.NoError(t, err)
	assert.Equal(t, int64(len(srcContent)), written)
	assert.Equal(t, srcContent, dst.Bytes())
}

func TestCopy_SingleLimiter(t *testing.T) {
	srcContent := []byte("test content")
	src := bytes.NewReader(srcContent)
	dst := &bytes.Buffer{}

	limiter := rate.NewLimiter(rate.Limit(10), 10)
	limiters := []*rate.Limiter{limiter}

	written, err := Copy(dst, src, limiters)
	require.NoError(t, err)
	assert.Equal(t, int64(len(srcContent)), written)
	assert.Equal(t, srcContent, dst.Bytes())
}

type ErrorReader struct{}

func (e *ErrorReader) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestCopy_Error(t *testing.T) {
	src := &ErrorReader{}
	dst := &bytes.Buffer{}

	limiters := []*rate.Limiter{}

	_, err := Copy(dst, src, limiters)
	assert.Error(t, err)
}
