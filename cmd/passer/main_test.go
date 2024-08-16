package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/galqiwi/fair-p/internal/testtool"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"sync"
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

func testProxyWithEchoService(t *testing.T, port string, echoService *httptest.Server) {
	proxy, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%s", port))
	require.NoError(t, err)

	msg := "hello world"

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	request, err := http.NewRequest(
		"POST",
		echoService.URL,
		bytes.NewBufferString(msg),
	)
	require.NoError(t, err)
	request.Header.Set("Content-Type", "text/plain")

	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, response.StatusCode)

	require.Equal(t, msg, string(body))
}

func testProxy(t *testing.T, testTLS bool, nRequests int) {
	echoHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, r.Body)
		defer func() { _ = r.Body.Close() }()
		w.Header().Set("Content-Length", fmt.Sprintf("%d", r.ContentLength))
	})

	var echoService *httptest.Server
	if testTLS {
		echoService = httptest.NewTLSServer(echoHandler)
	} else {
		echoService = httptest.NewServer(echoHandler)
	}
	defer echoService.Close()

	port, cleanup := startProxy(t)
	defer cleanup()

	var wg sync.WaitGroup

	wg.Add(nRequests)
	for i := 0; i < nRequests; i++ {
		go func() {
			defer wg.Done()
			testProxyWithEchoService(t, port, echoService)
		}()
	}

	wg.Wait()
}

func TestProxy(t *testing.T) {
	t.Run("http", func(t *testing.T) {
		testProxy(t, false, 1)
		testProxy(t, false, 4)
		// testProxy(t, false, 16)
	})
	t.Run("https", func(t *testing.T) {
		testProxy(t, true, 1)
		testProxy(t, true, 4)
		// testProxy(t, true, 16)
	})
}
