package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/galqiwi/fair-p/internal/utils"
	"net"
	"os"
	"time"
)

var concurrentConnectionsCounter = utils.NewCounter()

func Main() error {
	port := flag.String("port", "12345", "Port to listen on")
	proxyAddr := flag.String("proxy", "localhost:8080", "Address of the HTTP proxy")
	nRequests := flag.Int("n_requests", 1, "Number of requests")
	flag.Parse()

	address := fmt.Sprintf("localhost:%s", *port)

	// Listen for incoming connections.
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("error creating listener: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Server listening on %s\n", address)

	go func() {
		for {
			fmt.Printf("Concurrent connections: %d\n", concurrentConnectionsCounter.Get())
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for i := 0; i < *nRequests; i++ {
			go sendReq(address, *proxyAddr)
		}
	}()

	for {
		// Wait for a connection.
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection: %v\n", err)
			continue
		}

		// Handle the connection in a new goroutine.
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	concurrentConnectionsCounter.Add(1)
	defer concurrentConnectionsCounter.Sub(1)

	defer conn.Close()
	fmt.Printf("Client %s connected.\n", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		receivedText := scanner.Text()
		fmt.Printf("Received: %s\n", receivedText)

		// Echo the received text back to the client.
		_, err := conn.Write([]byte(receivedText + "\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to connection: %v\n", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from connection: %v\n", err)
	}

	fmt.Printf("Client %s disconnected.\n", conn.RemoteAddr().String())
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}
