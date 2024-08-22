package main

import (
	"bufio"
	"flag"
	"io"
	"net"
	"net/http"
	"net/url"
)

func main() {
	serverAddrPtr := flag.String("server", "", "Server address to send the request to")
	proxyAddrPtr := flag.String("proxy", "", "Proxy address to route the request through")
	numRequests := flag.Int("n_requests", 1, "Number of requests")
	flag.Parse()

	for i := 0; i < *numRequests; i++ {
		go sendReq(*serverAddrPtr, *proxyAddrPtr)
	}
}

func sendReq(serverAddr, proxyAddr string) {
	proxyConn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		panic(err)
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
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		panic(resp.StatusCode)
	}

	pipeReader, pipeWriter := io.Pipe()

	_ = pipeWriter
	//go func() {
	//	for {
	//		fmt.Fprintf(pipeWriter, "*\n")
	//		time.Sleep(time.Millisecond * 100)
	//	}
	//}()

	scanner := bufio.NewScanner(pipeReader)
	for scanner.Scan() {
		message := scanner.Text()
		_, err := proxyConn.Write([]byte(message + "\n"))
		if err != nil {
			panic(err)
		}

		responseScanner := bufio.NewScanner(proxyConn)
		if responseScanner.Scan() {
			_ = responseScanner.Text()
		}

		if err := responseScanner.Err(); err != nil {
			panic(err)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
