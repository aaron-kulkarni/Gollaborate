package main

import (
	"fmt"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const protocolID = protocol.ID("/p2p/1.0.0")
const serviceTag = "p2p-discovery"

func main() {
	// ctx := context.Background()

	// Create a new libp2p Host
	h, err := libp2p.New(
			libp2p.NATPortMap(), // Enable NAT traversal
			libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"), // Listen on all interfaces and a random port
		)

	if err != nil {
		log.Fatal(err)
	}

	// Set a stream handler on the host
	h.SetStreamHandler(protocolID, handleStream)

	// Print the host's multiaddresses
	fmt.Println("Main node is running. Connect to me at:")
	for _, addr := range h.Addrs() {
		fmt.Printf("%s/p2p/%s\n", addr, h.ID().String())
	}

	// Use mDNS for peer discovery
	mdnsService := mdns.NewMdnsService(h, serviceTag, &discoveryNotifee{h: h})
	if err != nil {
		log.Fatal(err)
	}
	defer mdnsService.Close()

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

type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	log.Println("Discovered new peer:", pi.ID.String())
	n.h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
}
