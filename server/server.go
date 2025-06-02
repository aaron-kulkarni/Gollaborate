package main

import (
	"fmt"
	"io"
	"net"
)

func main() {

	port := ":49874"

	// Listen on all interfaces on port
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	// Get the local IP address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Error getting local IP address:", err)
		return
	}

	var localIP string
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			localIP = ipNet.IP.String()
			break
		}
	}

	if localIP == "" {
		fmt.Println("Could not determine local IP address")
		return
	}

	addr := localIP + port

	fmt.Printf("Server is listening on %s\n", addr)
	fmt.Printf("Use the following command to run the client:\n")
	fmt.Printf("go run client.go %s\n", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024) // Buffer to read data in chunks
	for {
		// Read data from the client
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from client:", err)
			}
			return
		}

		// Print the received data
		fmt.Print(string(buf[:n])) // Print data as it's received
	}
}
