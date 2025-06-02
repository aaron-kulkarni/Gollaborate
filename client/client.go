package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"gollaborate/gui"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run client.go <server-ip>")
		return
	}

	serverAddress := os.Args[1]

	// Try to connect to server with retries
	var conn net.Conn
	var err error
	
	for attempts := 0; attempts < 3; attempts++ {
		conn, err = net.Dial("tcp", serverAddress)
		if err == nil {
			break
		}
		
		fmt.Printf("Connection attempt %d failed: %v\n", attempts+1, err)
		if attempts < 2 {
			fmt.Println("Retrying in 2 seconds...")
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		fmt.Printf("Failed to connect to server after 3 attempts: %v\n", err)
		fmt.Println("Starting in offline mode...")
		conn = nil
	} else {
		fmt.Printf("Connected to server at %s\n", serverAddress)
	}

	// Start the GUI with the connection (can be nil for offline mode)
	gui.Gui(conn)
	
	// Close connection when GUI exits
	if conn != nil {
		conn.Close()
	}
}