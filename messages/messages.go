package messages

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gollaborate/crdt"
	"net"
)

// MessageType represents the type of message being sent
type MessageType string

const (
	MessageTypeOperation MessageType = "operation"
	MessageTypeSync      MessageType = "sync"
	MessageTypeInit      MessageType = "init"
	MessageTypeAck       MessageType = "ack"
	MessageTypeError     MessageType = "error"
	MessageTypeCursor    MessageType = "cursor"
	MessageTypeSelection MessageType = "selection"
)

// OperationType represents the type of CRDT operation
type OperationType string

const (
	OperationTypeInsert OperationType = "insert"
	OperationTypeDelete OperationType = "delete"
)

// CursorPosition represents a cursor position using CRDT identifiers
type CursorPosition struct {
	Position []crdt.Identifier `json:"position"`
	UserID   int               `json:"user_id"`
	UserName string            `json:"user_name,omitempty"`
	Color    string            `json:"color,omitempty"` // Hex color for cursor display
}

// Selection represents a text selection range
type Selection struct {
	StartPosition []crdt.Identifier `json:"start_position"`
	EndPosition   []crdt.Identifier `json:"end_position"`
	UserID        int               `json:"user_id"`
	UserName      string            `json:"user_name,omitempty"`
	Color         string            `json:"color,omitempty"` // Hex color for selection display
}

// Operation represents a single CRDT operation
type Operation struct {
	Type      OperationType     `json:"type"`
	Position  []crdt.Identifier `json:"position"`
	Character rune              `json:"character,omitempty"`
	UserID    int               `json:"user_id"`
	Clock     int               `json:"clock"`
}

// Message represents a network message between client and server
type Message struct {
	Type      MessageType     `json:"type"`
	Operation *Operation      `json:"operation,omitempty"`
	Document  *crdt.Document  `json:"document,omitempty"`
	Cursor    *CursorPosition `json:"cursor,omitempty"`
	Selection *Selection      `json:"selection,omitempty"`
	UserID    int             `json:"user_id,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// Serialize converts a Message to JSON bytes
func (m *Message) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// Deserialize converts JSON bytes to a Message
func Deserialize(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// NewOperationMessage creates a new operation message
func NewOperationMessage(op *Operation) *Message {
	return &Message{
		Type:      MessageTypeOperation,
		Operation: op,
		UserID:    op.UserID,
	}
}

// NewSyncMessage creates a new sync message with the full document
func NewSyncMessage(doc *crdt.Document, userID int) *Message {
	return &Message{
		Type:     MessageTypeSync,
		Document: doc,
		UserID:   userID,
	}
}

// NewInitMessage creates a new init message for new client connections
func NewInitMessage(doc *crdt.Document) *Message {
	return &Message{
		Type:     MessageTypeInit,
		Document: doc,
	}
}

// NewAckMessage creates a new acknowledgment message
func NewAckMessage(userID int) *Message {
	return &Message{
		Type:   MessageTypeAck,
		UserID: userID,
	}
}

// NewErrorMessage creates a new error message
func NewErrorMessage(errorMsg string, userID int) *Message {
	return &Message{
		Type:   MessageTypeError,
		Error:  errorMsg,
		UserID: userID,
	}
}

// NewCursorMessage creates a new cursor position message
func NewCursorMessage(position []crdt.Identifier, userID int, userName, color string) *Message {
	return &Message{
		Type: MessageTypeCursor,
		Cursor: &CursorPosition{
			Position: position,
			UserID:   userID,
			UserName: userName,
			Color:    color,
		},
		UserID: userID,
	}
}

// NewSelectionMessage creates a new selection message
func NewSelectionMessage(startPos, endPos []crdt.Identifier, userID int, userName, color string) *Message {
	return &Message{
		Type: MessageTypeSelection,
		Selection: &Selection{
			StartPosition: startPos,
			EndPosition:   endPos,
			UserID:        userID,
			UserName:      userName,
			Color:         color,
		},
		UserID: userID,
	}
}

// NewInsertOperation creates a new insert operation
func NewInsertOperation(position []crdt.Identifier, character rune, userID int, clock int) *Operation {
	return &Operation{
		Type:      OperationTypeInsert,
		Position:  position,
		Character: character,
		UserID:    userID,
		Clock:     clock,
	}
}

// NewDeleteOperation creates a new delete operation
func NewDeleteOperation(position []crdt.Identifier, userID int, clock int) *Operation {
	return &Operation{
		Type:     OperationTypeDelete,
		Position: position,
		UserID:   userID,
		Clock:    clock,
	}
}

// SendMessage sends a message over a network connection
func SendMessage(conn net.Conn, msg *Message) error {
	data, err := msg.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}
	
	// Add newline delimiter for easier parsing
	data = append(data, '\n')
	
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	
	return nil
}

// ReceiveMessage receives a message from a network connection
func ReceiveMessage(conn net.Conn) (*Message, error) {
	reader := bufio.NewReader(conn)
	
	// Read until newline delimiter
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}
	
	// Remove the newline delimiter
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	
	msg, err := Deserialize(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %w", err)
	}
	
	return msg, nil
}

// SendOperation is a convenience function to send an operation message
func SendOperation(conn net.Conn, op *Operation) error {
	msg := NewOperationMessage(op)
	return SendMessage(conn, msg)
}

// SendSync is a convenience function to send a sync message
func SendSync(conn net.Conn, doc *crdt.Document, userID int) error {
	msg := NewSyncMessage(doc, userID)
	return SendMessage(conn, msg)
}

// SendInit is a convenience function to send an init message
func SendInit(conn net.Conn, doc *crdt.Document) error {
	msg := NewInitMessage(doc)
	return SendMessage(conn, msg)
}

// SendError is a convenience function to send an error message
func SendError(conn net.Conn, errorMsg string, userID int) error {
	msg := NewErrorMessage(errorMsg, userID)
	return SendMessage(conn, msg)
}

// SendCursor is a convenience function to send a cursor position message
func SendCursor(conn net.Conn, position []crdt.Identifier, userID int, userName, color string) error {
	msg := NewCursorMessage(position, userID, userName, color)
	return SendMessage(conn, msg)
}

// SendSelection is a convenience function to send a selection message
func SendSelection(conn net.Conn, startPos, endPos []crdt.Identifier, userID int, userName, color string) error {
	msg := NewSelectionMessage(startPos, endPos, userID, userName, color)
	return SendMessage(conn, msg)
}

// SendClearSelection sends an empty selection to clear a user's selection
func SendClearSelection(conn net.Conn, userID int, userName, color string) error {
	msg := NewSelectionMessage(nil, nil, userID, userName, color)
	return SendMessage(conn, msg)
}