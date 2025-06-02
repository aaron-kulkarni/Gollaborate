package main

import (
	"bytes"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"gollaborate/crdt"
	"gollaborate/messages"
)

// PeerSim simulates a decentralized peer with in-memory connections.
type PeerSim struct {
	ID       int
	Doc      *crdt.Document
	Ops      []*messages.Operation
	Peers    []*PeerSim
	RecvChan chan *messages.Operation
	Mutex    sync.Mutex
}

func NewPeerSim(id int) *PeerSim {
	return &PeerSim{
		ID:       id,
		Doc:      crdt.FromText("", id),
		Ops:      []*messages.Operation{},
		RecvChan: make(chan *messages.Operation, 100),
	}
}

// Connects two peers bidirectionally.
func ConnectPeers(a, b *PeerSim) {
	a.Peers = append(a.Peers, b)
	b.Peers = append(b.Peers, a)
}

// Broadcasts an operation to all connected peers.
func (p *PeerSim) Broadcast(op *messages.Operation) {
	for _, peer := range p.Peers {
		peer.RecvChan <- op
	}
}

// Processes incoming operations.
func (p *PeerSim) Run(wg *sync.WaitGroup, stop <-chan struct{}) {
	defer wg.Done()
	for {
		select {
		case op := <-p.RecvChan:
			p.Mutex.Lock()
			_ = p.Doc.InsertCharacter(op.Character, op.Position, op.Clock)
			p.Ops = append(p.Ops, op)
			p.Mutex.Unlock()
		case <-stop:
			return
		}
	}
}

// Simulates a local edit and broadcasts it.
func (p *PeerSim) LocalEdit(char rune, clock int) *messages.Operation {
	pos, _ := p.Doc.GeneratePositionAt(1, len(p.Doc.ToText())+1, p.ID)
	op := messages.NewInsertOperation(pos, char, p.ID, clock)
	_ = p.Doc.InsertCharacter(char, pos, clock)
	p.Ops = append(p.Ops, op)
	p.Broadcast(op)
	return op
}

func TestPeerToPeerPropagation(t *testing.T) {
	peerA := NewPeerSim(1)
	peerB := NewPeerSim(2)
	peerC := NewPeerSim(3)
	ConnectPeers(peerA, peerB)
	ConnectPeers(peerB, peerC)
	ConnectPeers(peerA, peerC)

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(3)
	go peerA.Run(&wg, stop)
	go peerB.Run(&wg, stop)
	go peerC.Run(&wg, stop)

	// Peer A inserts 'X'
	peerA.LocalEdit('X', 1)
	time.Sleep(50 * time.Millisecond)

	peerB.LocalEdit('Y', 2)
	time.Sleep(50 * time.Millisecond)

	peerC.LocalEdit('Z', 3)
	time.Sleep(100 * time.Millisecond)

	close(stop)
	wg.Wait()

	// All peers should have all operations (order may differ due to CRDT)
	for _, peer := range []*PeerSim{peerA, peerB, peerC} {
		text := peer.Doc.ToText()
		if !containsAll(text, []rune{'X', 'Y', 'Z'}) {
			t.Errorf("Peer %d document missing characters: got '%s'", peer.ID, text)
		}
	}
	// All peers should have the same document state (modulo CRDT order)
	if !crdtDocsEquivalent(peerA.Doc, peerB.Doc) || !crdtDocsEquivalent(peerA.Doc, peerC.Doc) {
		t.Errorf("Peers' documents diverged: '%s', '%s', '%s'", peerA.Doc.ToText(), peerB.Doc.ToText(), peerC.Doc.ToText())
	}
}

func TestPeerJoinAndSync(t *testing.T) {
	peerA := NewPeerSim(1)
	peerB := NewPeerSim(2)
	ConnectPeers(peerA, peerB)

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go peerA.Run(&wg, stop)
	go peerB.Run(&wg, stop)

	peerA.LocalEdit('A', 1)
	peerA.LocalEdit('B', 2)
	time.Sleep(50 * time.Millisecond)

	// New peer joins and syncs from peerA
	peerC := NewPeerSim(3)
	ConnectPeers(peerA, peerC)
	wg.Add(1)
	go peerC.Run(&wg, stop)

	// Simulate sync: peerA sends its document state to peerC
	docBytes, _ := json.Marshal(peerA.Doc)
	var docCopy crdt.Document
	_ = json.Unmarshal(docBytes, &docCopy)
	peerC.Doc = &docCopy

	// PeerC edits
	opC := peerC.LocalEdit('C', 3)
	// Manually broadcast C's operation to all peers (simulate gossip)
	for _, peer := range []*PeerSim{peerA, peerB} {
		peer.RecvChan <- opC
	}
	time.Sleep(100 * time.Millisecond)

	close(stop)
	wg.Wait()

	for _, peer := range []*PeerSim{peerA, peerB, peerC} {
		text := peer.Doc.ToText()
		if !containsAll(text, []rune{'A', 'B', 'C'}) {
			t.Errorf("Peer %d document missing characters after join: got '%s'", peer.ID, text)
		}
	}
}

func TestConcurrentEdits(t *testing.T) {
	peerA := NewPeerSim(1)
	peerB := NewPeerSim(2)
	ConnectPeers(peerA, peerB)

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go peerA.Run(&wg, stop)
	go peerB.Run(&wg, stop)

	// Simultaneous edits
	go peerA.LocalEdit('Q', 1)
	go peerB.LocalEdit('W', 1)
	time.Sleep(100 * time.Millisecond)

	close(stop)
	wg.Wait()

	textA := peerA.Doc.ToText()
	textB := peerB.Doc.ToText()
	if !containsAll(textA, []rune{'Q', 'W'}) || !containsAll(textB, []rune{'Q', 'W'}) {
		t.Errorf("Concurrent edits not merged: '%s', '%s'", textA, textB)
	}
	if !crdtDocsEquivalent(peerA.Doc, peerB.Doc) {
		t.Errorf("Peers' documents diverged after concurrent edits: '%s', '%s'", textA, textB)
	}
}

// Helper: checks all runes are present in s
func containsAll(s string, chars []rune) bool {
	for _, c := range chars {
		found := false
		for _, sc := range s {
			if sc == c {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Helper: checks if two CRDT documents are equivalent (by text content)
func crdtDocsEquivalent(a, b *crdt.Document) bool {
	return a.ToText() == b.ToText()
}

// --- Mock net.Conn for completeness (not used in these tests, but for future expansion) ---

type MockConn struct {
	buf    bytes.Buffer
	closed bool
}

func (m *MockConn) Read(b []byte) (int, error)         { return m.buf.Read(b) }
func (m *MockConn) Write(b []byte) (int, error)        { return m.buf.Write(b) }
func (m *MockConn) Close() error                       { m.closed = true; return nil }
func (m *MockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (m *MockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }
