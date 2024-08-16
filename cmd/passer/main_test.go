package main

import (
	"bytes"
	"fmt"
	"github.com/galqiwi/fair-p/internal/testtool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"testing"
	"time"
)

const importPath = "github.com/galqiwi/fair-p/cmd/passer"

var binCache testtool.BinCache

func TestMain(m *testing.M) {
	os.Exit(func() int {
		var teardown testtool.CloseFunc
		binCache, teardown = testtool.NewBinCache()
		defer teardown()

		return m.Run()
	}())
}

func startProxy(t *testing.T) (port string, stop func()) {
	binary, err := binCache.GetBinary(importPath)
	require.NoError(t, err)

	port, err = testtool.GetFreePort()
	require.NoError(t, err, "unable to get free port")

	cmd := exec.Command(binary, "--port", port, "--max_throughput", "1")
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr

	require.NoError(t, cmd.Start())

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	stop = func() {
		_ = cmd.Process.Kill()
		<-done
	}

	if err = testtool.WaitForPort(t, time.Second*5, port); err != nil {
		stop()
	}

	require.NoError(t, err)
	return
}

func TestProxy(t *testing.T) {
	echoService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, r.Body)
		defer func() { _ = r.Body.Close() }()
		w.Header().Set("Content-Length", fmt.Sprintf("%d", r.ContentLength))
	}))
	defer echoService.Close()

	port, cleanup := startProxy(t)
	defer cleanup()

	proxy, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%s", port))
	assert.NoError(t, err)

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}

	msg := "hello world"

	request, err := http.NewRequest(
		"POST",
		echoService.URL,
		bytes.NewBufferString(msg),
	)
	assert.NoError(t, err)
	request.Header.Set("Content-Type", "text/plain")

	response, err := client.Do(request)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	assert.NoError(t, err)

	assert.Equal(t, msg, string(body))
}
