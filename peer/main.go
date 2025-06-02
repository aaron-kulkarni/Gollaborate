package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"gollaborate/gui"
)

type Peer struct {
	ID         int
	ListenAddr string
	Peers      map[string]net.Conn
	Mutex      sync.Mutex
}

func generatePeerID() int {
	// Use current time and random data for a simple unique ID
	return int(time.Now().UnixNano() % 99999999)
}

func (p *Peer) connectToPeer(addr string, editorState *gui.EditorState) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	if _, exists := p.Peers[addr]; exists {
		return
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to connect to peer %s: %v\n", addr, err)
		return
	}
	p.Peers[addr] = conn
	fmt.Printf("Connected to peer: %s\n", addr)
	if editorState != nil {
		editorState.AddConn(conn)
	}
}

func (p *Peer) listenForPeers(editorState *gui.EditorState) {
	ln, err := net.Listen("tcp", p.ListenAddr)
	if err != nil {
		fmt.Printf("Failed to listen on %s: %v\n", p.ListenAddr, err)
		os.Exit(1)
	}
	fmt.Printf("Listening for peers on %s\n", p.ListenAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		remoteAddr := conn.RemoteAddr().String()
		p.Mutex.Lock()
		p.Peers[remoteAddr] = conn
		p.Mutex.Unlock()
		fmt.Printf("Accepted connection from peer: %s\n", remoteAddr)
		if editorState != nil {
			editorState.AddConn(conn)
		}
	}
}

func main() {
	listenAddr := flag.String("listen", "0.0.0.0:49874", "Address to listen for incoming peer connections")
	peerList := flag.String("peers", "", "Comma-separated list of peer addresses to connect to")
	flag.Parse()

	peer := &Peer{
		ID:         generatePeerID(),
		ListenAddr: *listenAddr,
		Peers:      make(map[string]net.Conn),
	}

	// Create the editor state up front so connections can be added dynamically
	editorState := gui.NewEditorState(nil, peer.ID)

	// Start listening for incoming peers, passing the editor state
	go peer.listenForPeers(editorState)

	// Connect to peers specified in the command line
	if *peerList != "" {
		peers := strings.Split(*peerList, ",")
		for _, addr := range peers {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				go peer.connectToPeer(addr, editorState)
			}
		}
	}

	// Start the GUI with the editor state (must be on main goroutine)
	gui.GuiWithPeers(nil, peer.ID, editorState)
}
