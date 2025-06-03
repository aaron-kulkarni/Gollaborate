package shared

import (
	"net"
	"sync"

	"gollaborate/crdt"
	"gollaborate/messages"
)

// MessageListener is a function that receives messages
type MessageListener func(*messages.Message)

type EditorState struct {
	document   *crdt.Document
	nodeID     int
	conns      []net.Conn
	mutex      sync.Mutex
	listeners  []MessageListener
	currentClock int
}

// For testing purposes
func (e *EditorState) SetDocument(doc *crdt.Document) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.document = doc
}

func NewEditorState(doc *crdt.Document, nodeID int) *EditorState {
	return &EditorState{
		document:   doc,
		nodeID:     nodeID,
		conns:      []net.Conn{},
		listeners:  []MessageListener{},
		currentClock: 1,
	}
}

func (e *EditorState) Document() *crdt.Document {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.document
}

func (e *EditorState) NodeID() int {
	return e.nodeID
}

func (e *EditorState) AddConn(conn net.Conn) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.conns = append(e.conns, conn)
	
	// Start listening for messages from this connection
	go e.listenForMessages(conn)
}

func (e *EditorState) Connections() []net.Conn {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	// Return a copy to avoid concurrent modification issues
	connsCopy := make([]net.Conn, len(e.conns))
	copy(connsCopy, e.conns)
	return connsCopy
}

// AddMessageListener adds a function to be called when a message is received
func (e *EditorState) AddMessageListener(listener MessageListener) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.listeners = append(e.listeners, listener)
}

// BroadcastMessage sends a message to all connected peers
func (e *EditorState) BroadcastMessage(msg *messages.Message) {
	conns := e.Connections()
	for _, conn := range conns {
		err := messages.SendMessage(conn, msg)
		if err != nil {
			// Handle error, maybe remove the connection
			e.removeConnection(conn)
		}
	}
}

// InsertCharacter inserts a character into the document and broadcasts the operation
func (e *EditorState) InsertCharacter(char rune, pos []crdt.Identifier) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	// Update local clock
	e.currentClock++
	clock := e.currentClock
	
	// Apply to local document
	err := e.document.InsertCharacter(char, pos, clock)
	if err != nil {
		return err
	}
	
	// Create and broadcast operation
	op := messages.NewInsertOperation(pos, char, e.nodeID, clock)
	msg := messages.NewOperationMessage(op)
	
	go e.BroadcastMessage(msg)
	return nil
}

// DeleteCharacter deletes a character from the document and broadcasts the operation
func (e *EditorState) DeleteCharacter(pos []crdt.Identifier) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	// Update local clock
	e.currentClock++
	clock := e.currentClock
	
	// Apply to local document
	err := e.document.DeleteCharacter(pos)
	if err != nil {
		return err
	}
	
	// Create and broadcast operation
	op := messages.NewDeleteOperation(pos, e.nodeID, clock)
	msg := messages.NewOperationMessage(op)
	
	go e.BroadcastMessage(msg)
	return nil
}

// SyncDocument sends the current document state to all peers
func (e *EditorState) SyncDocument() {
	e.mutex.Lock()
	doc := e.document
	e.mutex.Unlock()
	
	msg := messages.NewSyncMessage(doc, e.nodeID)
	go e.BroadcastMessage(msg)
}

// listenForMessages continuously listens for messages from a connection
func (e *EditorState) listenForMessages(conn net.Conn) {
	for {
		msg, err := messages.ReceiveMessage(conn)
		if err != nil {
			// Connection likely closed
			e.removeConnection(conn)
			return
		}
		
		// Handle the message
		e.handleMessage(msg)
	}
}

// handleMessage processes incoming messages and updates state
func (e *EditorState) handleMessage(msg *messages.Message) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	switch msg.Type {
	case messages.MessageTypeOperation:
		if msg.Operation != nil && msg.Operation.UserID != e.nodeID {
			op := msg.Operation
			switch op.Type {
			case messages.OperationTypeInsert:
				_ = e.document.InsertCharacter(op.Character, op.Position, op.Clock)
			case messages.OperationTypeDelete:
				_ = e.document.DeleteCharacter(op.Position)
			}
		}
	case messages.MessageTypeSync:
		if msg.Document != nil && msg.UserID != e.nodeID {
			e.document = msg.Document
		}
	}
	
	// Notify listeners
	for _, listener := range e.listeners {
		go listener(msg)
	}
}

// removeConnection removes a connection from the connection list
func (e *EditorState) removeConnection(conn net.Conn) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	for i, c := range e.conns {
		if c == conn {
			// Close connection if not already closed
			_ = conn.Close()
			// Remove from slice
			e.conns = append(e.conns[:i], e.conns[i+1:]...)
			break
		}
	}
}