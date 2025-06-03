package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gollaborate/crdt"
	"gollaborate/messages"
	"gollaborate/shared"
	"gollaborate/tui/core"
)

var (
	port      = flag.Int("port", 8080, "Port to listen on")
	nodeID    = flag.Int("node", 0, "Node ID (0 for random)")
	join      = flag.String("join", "", "Address of node to join (host:port)")
	textFile  = flag.String("file", "", "Text file to load (optional)")
	username  = flag.String("user", "", "Username (optional)")
	colorName = flag.String("color", "blue", "User color (blue, green, red, yellow, cyan, magenta)")
)

// Available colors for users
var colors = map[string]string{
	"blue":    "34",
	"green":   "32",
	"red":     "31",
	"yellow":  "33",
	"cyan":    "36",
	"magenta": "35",
}

func main() {
	flag.Parse()

	// Generate random node ID if not specified
	userNodeID := *nodeID
	if userNodeID == 0 {
		rand.Seed(time.Now().UnixNano())
		userNodeID = rand.Intn(999) + 1
	}

	// Set username if not specified
	user := *username
	if user == "" {
		user = fmt.Sprintf("User-%d", userNodeID)
	}

	// Validate color
	color, ok := colors[*colorName]
	if !ok {
		color = colors["blue"]
	}

	// Initialize document
	var doc *crdt.Document
	if *textFile != "" {
		// Try to load document from file
		content, err := os.ReadFile(*textFile)
		if err != nil {
			log.Printf("Failed to load file %s: %v, starting with empty document", *textFile, err)
			doc = crdt.FromText("", userNodeID)
		} else {
			doc = crdt.FromText(string(content), userNodeID)
			log.Printf("Loaded document from %s", *textFile)
		}
	} else {
		// Start with empty document
		doc = crdt.FromText("", userNodeID)
		log.Printf("Starting with empty document")
	}

	// Create editor state
	editorState := shared.NewEditorState(doc, userNodeID)

	// Setup network listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()
	log.Printf("Listening on port %d", *port)

	// Handle incoming connections in a goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}
			log.Printf("New connection from %s", conn.RemoteAddr())

			// Add connection to editor state
			editorState.AddConn(conn)

			// Send current document state to new peer
			err = messages.SendSync(conn, editorState.Document(), userNodeID)
			if err != nil {
				log.Printf("Error sending document sync: %v", err)
			}
		}
	}()

	// Join existing network if specified
	if *join != "" {
		log.Printf("Attempting to join %s...", *join)
		conn, err := net.Dial("tcp", *join)
		if err != nil {
			log.Printf("Failed to connect to %s: %v", *join, err)
		} else {
			log.Printf("Connected to %s", *join)
			editorState.AddConn(conn)

			// Request document sync
			err = messages.SendInit(conn, nil, userNodeID)
			if err != nil {
				log.Printf("Error requesting document sync: %v", err)
			}
		}
	}

	// Handle signals for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down...")

		// Save document if file was specified
		if *textFile != "" {
			text := editorState.Document().ToText()
			err := os.WriteFile(*textFile, []byte(text), 0644)
			if err != nil {
				log.Printf("Error saving document: %v", err)
			} else {
				log.Printf("Document saved to %s", *textFile)
			}
		}

		os.Exit(0)
	}()

	// Start TUI
	log.Printf("Starting Gollaborate TUI as node %d", userNodeID)
	if err := core.StartTUI(editorState, userNodeID, color); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}
