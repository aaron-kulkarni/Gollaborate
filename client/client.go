package main

import (
	"fmt"
	"net"
	"os"

	"gollaborate/gui"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run client.go <server-ip>")
		return
	}

	serverAddress := os.Args[1]

	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	gui.Gui(conn)
	defer conn.Close()

}
