package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"gollaborate/crdt"
	"gollaborate/messages"
	"gollaborate/users"
)

// Server manages the collaborative document and connected clients
type Server struct {
	document    *crdt.Document
	userManager *users.Manager
	clients     map[int]*Client
	mutex       sync.RWMutex
	nodeID      int
	clock       int
}

// Client represents a connected client
type Client struct {
	ID     int
	User   *users.User
	Conn   net.Conn
	Server *Server
}

// NewServer creates a new collaborative server
func NewServer() *Server {
	return &Server{
		document:    crdt.FromText("", 0), // Server uses node ID 0
		userManager: users.NewManager(),
		clients:     make(map[int]*Client),
		nodeID:      0,
		clock:       1,
	}
}

func (s *Server) nextClock() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clock++
	return s.clock
}

// AddClient adds a new client to the server
func (s *Server) AddClient(conn net.Conn) *Client {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	user := s.userManager.CreateUser(fmt.Sprintf("User%d", s.userManager.GetNextAvailableID()))
	client := &Client{
		ID:     user.ID,
		User:   user,
		Conn:   conn,
		Server: s,
	}

	s.clients[client.ID] = client
	log.Printf("Client %d (%s) connected from %s", client.ID, client.User.Name, conn.RemoteAddr())

	// Send initial document state to the new client
	messages.SendInit(conn, s.document)

	return client
}

// RemoveClient removes a client from the server
func (s *Server) RemoveClient(clientID int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if client, exists := s.clients[clientID]; exists {
		log.Printf("Client %d (%s) disconnected", clientID, client.User.Name)
		client.Conn.Close()
		delete(s.clients, clientID)
		s.userManager.RemoveUser(clientID)
	}
}

// BroadcastOperation sends an operation to all clients except the sender
func (s *Server) BroadcastOperation(senderID int, operation *messages.Operation) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for clientID, client := range s.clients {
		if clientID != senderID {
			err := messages.SendOperation(client.Conn, operation)
			if err != nil {
				log.Printf("Error sending operation to client %d: %v", clientID, err)
				// Don't remove client here to avoid deadlock, mark for cleanup
				go s.RemoveClient(clientID)
			}
		}
	}
}

// BroadcastCursor sends cursor position to all clients except the sender
func (s *Server) BroadcastCursor(senderID int, cursor *messages.CursorPosition) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for clientID, client := range s.clients {
		if clientID != senderID {
			err := messages.SendCursor(client.Conn, cursor.Position, cursor.UserID, cursor.UserName, cursor.Color)
			if err != nil {
				log.Printf("Error sending cursor to client %d: %v", clientID, err)
				go s.RemoveClient(clientID)
			}
		}
	}
}

// BroadcastSelection sends selection to all clients except the sender
func (s *Server) BroadcastSelection(senderID int, selection *messages.Selection) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for clientID, client := range s.clients {
		if clientID != senderID {
			err := messages.SendSelection(client.Conn, selection.StartPosition, selection.EndPosition, selection.UserID, selection.UserName, selection.Color)
			if err != nil {
				log.Printf("Error sending selection to client %d: %v", clientID, err)
				go s.RemoveClient(clientID)
			}
		}
	}
}

// ApplyOperation applies an operation to the server's document
func (s *Server) ApplyOperation(operation *messages.Operation) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Update server clock
	if operation.Clock > s.clock {
		s.clock = operation.Clock
	}

	// Apply operation to server document
	switch operation.Type {
	case messages.OperationTypeInsert:
		return s.document.InsertCharacter(operation.Character, operation.Position, operation.Clock)
	case messages.OperationTypeDelete:
		return s.document.DeleteCharacter(operation.Position)
	default:
		return fmt.Errorf("unknown operation type: %s", operation.Type)
	}
}

// GetDocumentState returns the current document state
func (s *Server) GetDocumentState() *crdt.Document {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.document
}

// GetConnectedUsers returns a list of connected users
func (s *Server) GetConnectedUsers() []*users.User {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var connectedUsers []*users.User
	for _, client := range s.clients {
		connectedUsers = append(connectedUsers, client.User)
	}
	return connectedUsers
}

// HandleClient processes messages from a client
func (c *Client) HandleClient() {
	defer c.Server.RemoveClient(c.ID)

	for {
		msg, err := messages.ReceiveMessage(c.Conn)
		if err != nil {
			log.Printf("Error receiving message from client %d: %v", c.ID, err)
			return
		}

		switch msg.Type {
		case messages.MessageTypeOperation:
			err := c.handleOperation(msg.Operation)
			if err != nil {
				log.Printf("Error handling operation from client %d: %v", c.ID, err)
				messages.SendError(c.Conn, err.Error(), c.ID)
			}

		case messages.MessageTypeCursor:
			c.handleCursor(msg.Cursor)

		case messages.MessageTypeSelection:
			c.handleSelection(msg.Selection)

		case messages.MessageTypeSync:
			c.handleSync()

		default:
			log.Printf("Unknown message type from client %d: %s", c.ID, msg.Type)
		}
	}
}

func (c *Client) handleOperation(operation *messages.Operation) error {
	if operation == nil {
		return fmt.Errorf("received nil operation")
	}

	// Validate operation
	if operation.UserID != c.ID {
		return fmt.Errorf("operation user ID %d doesn't match client ID %d", operation.UserID, c.ID)
	}

	// Apply operation to server document
	err := c.Server.ApplyOperation(operation)
	if err != nil {
		return fmt.Errorf("failed to apply operation: %w", err)
	}

	// Broadcast operation to other clients
	c.Server.BroadcastOperation(c.ID, operation)

	log.Printf("Applied %s operation from client %d (%s)", operation.Type, c.ID, c.User.Name)
	return nil
}

func (c *Client) handleCursor(cursor *messages.CursorPosition) {
	if cursor == nil {
		return
	}

	// Update cursor with user info
	cursor.UserID = c.ID
	cursor.UserName = c.User.Name
	cursor.Color = c.User.Color

	// Broadcast cursor position to other clients
	c.Server.BroadcastCursor(c.ID, cursor)
}

func (c *Client) handleSelection(selection *messages.Selection) {
	if selection == nil {
		return
	}

	// Update selection with user info
	selection.UserID = c.ID
	selection.UserName = c.User.Name
	selection.Color = c.User.Color

	// Broadcast selection to other clients
	c.Server.BroadcastSelection(c.ID, selection)
}

func (c *Client) handleSync() {
	// Send current document state to client
	doc := c.Server.GetDocumentState()
	messages.SendSync(c.Conn, doc, c.Server.nodeID)
	log.Printf("Sent sync to client %d (%s)", c.ID, c.User.Name)
}

func main() {
	port := ":49874"

	// Create server instance
	server := NewServer()

	// Listen on all interfaces
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer ln.Close()

	// Get local IP address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("Error getting local IP address: %v", err)
	}

	var localIP string
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			localIP = ipNet.IP.String()
			break
		}
	}

	if localIP == "" {
		localIP = "localhost"
	}

	addr := localIP + port
	log.Printf("Collaborative server started on %s", addr)
	log.Printf("Document initialized. Waiting for clients...")
	log.Printf("Use: go run client.go %s", addr)

	// Accept client connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		// Create new client and start handling
		client := server.AddClient(conn)
		go client.HandleClient()
	}
}
