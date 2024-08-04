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

	logger.Info("got request",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("host", r.Host),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.String("headers", strings.Join(headers, ", ")),
	)

	if r.Method == http.MethodConnect {
		handleTunneling(w, r)
	} else {
		handleHTTP(w, r)
	}
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go utils.Transfer(dest_conn, client_conn)
	go utils.Transfer(client_conn, dest_conn)
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	utils.CopyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
