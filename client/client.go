package main

import (
	"context"
	"fmt"
	"log"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

const protocolID = protocol.ID("/p2p/1.0.0")

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run client.go <main-node-multiaddr>")
		return
	}

	mainNodeAddr := os.Args[1]

	ctx := context.Background()

	// Create a new libp2p Host
	// h, err := libp2p.New(ctx)
	h, err := libp2p.New(
		libp2p.NATPortMap(),                            // Enable NAT traversal
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"), // Listen on all interfaces and a random port
	)
	if err != nil {
		log.Fatal(err)
	}

	// Set a stream handler on the host
	h.SetStreamHandler(protocolID, handleStream)

	// Parse the main node's multiaddress
	maddr, err := multiaddr.NewMultiaddr(mainNodeAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Extract the peer ID from the multiaddress
	peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the main node
	if err := h.Connect(ctx, *peerInfo); err != nil {
		log.Fatal(err)
	}

	// Open a stream to the main node
	s, err := h.NewStream(ctx, peerInfo.ID, protocolID)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	// Send a message to the main node
	message := "Hello from peer!"
	_, err = s.Write([]byte(message))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Message sent to main node:", message)

	select {}
}

func handleStream(s network.Stream) {
	log.Println("Got a new stream!")
	defer s.Close()

	buf := make([]byte, 1024)
	n, err := s.Read(buf)
	if err != nil {
		log.Println("Error reading from stream:", err)
		return
	}

	log.Printf("Received: %s\n", string(buf[:n]))
}
