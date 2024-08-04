package main

import (
	"crypto/tls"
	"fmt"
	"github.com/galqiwi/fair-p/internal/utils"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = utils.NewLogger()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {
	go func() {
		for {
			utils.LogRuntimeInfo(logger)
			time.Sleep(time.Second * 10)
		}
	}()
}

func main() {
	err := Main()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func Main() error {
	args := getArgs()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", args.port),
		Handler: http.HandlerFunc(mainHandler),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	return server.ListenAndServe()
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
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
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.String("headers", strings.Join(headers, ", ")),
	)

	if r.Method == http.MethodConnect {
		handleTunneling(w, r)
		return
	}

	if strings.HasPrefix(r.URL.String(), "/register") {
		logger.Info("Registered host", zap.String("url", r.URL.String()))
		fmt.Fprintf(w, "ok, thanks :)\n")
		return
	}

	handleHTTP(w, r)
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		logger.Info("Error dialing destination", zap.String("host", r.Host), zap.String("err", err.Error()))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		logger.Info("Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		logger.Info("Hijacking error", zap.String("err", err.Error()))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	logger.Info("Tunnel established", zap.String("client", r.RemoteAddr), zap.String("destination", r.Host))
	go utils.Transfer(destConn, clientConn)
	go utils.Transfer(clientConn, destConn)
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		logger.Info("RoundTrip error", zap.String("url", req.URL.String()), zap.String("err", err.Error()))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	utils.CopyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	bytesCopied, err := io.Copy(w, resp.Body)

	if err != nil {
		logger.Info("Error copying response body", zap.String("url", req.URL.String()), zap.String("err", err.Error()))
		return
	}
	logger.Info("HTTP response forwarded", zap.String("url", req.URL.String()), zap.Int64("bytes_copied", bytesCopied))
}
