package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

func sendReq(serverAddr, proxyAddr string) {
	proxyConn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to proxy: %v\n", err)
		os.Exit(1)
	}
	defer proxyConn.Close()

	connectReq := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: serverAddr},
		Host:   serverAddr,
		Header: make(http.Header),
	}

	connectReq.Write(proxyConn)

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), connectReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response from proxy: %v\n", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Proxy returned non-200 status: %v\n", resp.Status)
		os.Exit(1)
	}

	pipeReader, pipeWriter := io.Pipe()

	go func() {
		for {
			fmt.Fprintf(pipeWriter, "*")
			time.Sleep(time.Millisecond * 100)
		}
	}()

	scanner := bufio.NewScanner(pipeReader)
	for scanner.Scan() {
		message := scanner.Text()
		_, err := proxyConn.Write([]byte(message + "\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to connection: %v\n", err)
			os.Exit(1)
		}

		responseScanner := bufio.NewScanner(proxyConn)
		if responseScanner.Scan() {
			response := responseScanner.Text()
			fmt.Printf("Received: %s\n", response)
		}

		if err := responseScanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from connection: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
	}
}
