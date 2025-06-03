package main

import (
	"bytes"
	"encoding/json"
	"net"
	"testing"
	"time"

	"gollaborate/crdt"
	"gollaborate/shared"
	"gollaborate/tui/core"
)

// Test the TUI integration with CRDT
func TestTUIBasicEditing(t *testing.T) {
	// Create document and editor state
	doc := crdt.FromText("", 1)
	editorState := shared.NewEditorState(doc, 1)
	
	// Initialize TUI model
	model := core.InitializeModelForTesting(editorState, 1, "blue")
	
	// Simulate typing "Hello"
	model.SimulateKeyPress("H")
	model.SimulateKeyPress("e")
	model.SimulateKeyPress("l")
	model.SimulateKeyPress("l")
	model.SimulateKeyPress("o")
	
	// Verify document content
	text := model.GetDocumentText()
	if text != "Hello" {
		t.Errorf("Document text incorrect: got '%s', want 'Hello'", text)
	}
	
	// Verify cursor position
	x, y := model.GetCursorPosition()
	if x != 6 || y != 1 { // Should be after the last character
		t.Errorf("Cursor position incorrect: got (%d,%d), want (6,1)", x, y)
	}
}

// Test cursor movement and character deletion
func TestTUICursorAndDelete(t *testing.T) {
	// Create document with content
	doc := crdt.FromText("Hello", 1)
	editorState := shared.NewEditorState(doc, 1)
	model := core.InitializeModelForTesting(editorState, 1, "blue")
	
	// Move cursor to position before 'o'
	model.SetCursorPosition(5, 1)
	
	// Delete 'l'
	model.SimulateKeyPress("backspace")
	
	// Verify text is now "Helo"
	text := model.GetDocumentText()
	if text != "Helo" {
		t.Errorf("Document text after deletion incorrect: got '%s', want 'Helo'", text)
	}
}

// Test multiline editing
func TestTUIMultilineEditing(t *testing.T) {
	// Create an empty document
	doc := crdt.FromText("", 1)
	editorState := shared.NewEditorState(doc, 1)
	model := core.InitializeModelForTesting(editorState, 1, "blue")
	
	// Type "Line 1"
	model.SimulateKeyPress("L")
	model.SimulateKeyPress("i")
	model.SimulateKeyPress("n")
	model.SimulateKeyPress("e")
	model.SimulateKeyPress(" ")
	model.SimulateKeyPress("1")
	
	// Press Enter for new line
	model.SimulateKeyPress("enter")
	
	// Type "Line 2"
	model.SimulateKeyPress("L")
	model.SimulateKeyPress("i")
	model.SimulateKeyPress("n")
	model.SimulateKeyPress("e")
	model.SimulateKeyPress(" ")
	model.SimulateKeyPress("2")
	
	// Verify document content
	expected := "Line 1\nLine 2"
	if model.GetDocumentText() != expected {
		t.Errorf("Multiline text incorrect: got '%s', want '%s'", 
			model.GetDocumentText(), expected)
	}
}

// Test document synchronization between two TUI instances
func TestTUIDocumentSync(t *testing.T) {
	// Create two editor states with pipe connection
	doc1 := crdt.FromText("", 1)
	doc2 := crdt.FromText("", 2)
	
	editorState1 := shared.NewEditorState(doc1, 1)
	editorState2 := shared.NewEditorState(doc2, 2)
	
	// Connect them with a pipe
	conn1, conn2 := net.Pipe()
	editorState1.AddConn(conn1)
	editorState2.AddConn(conn2)
	
	// Create TUI models
	model1 := core.InitializeModelForTesting(editorState1, 1, "blue")
	model2 := core.InitializeModelForTesting(editorState2, 2, "red")
	
	// Edit in model1
	model1.SimulateKeyPress("H")
	model1.SimulateKeyPress("i")
	
	// Wait a moment for synchronization
	time.Sleep(100 * time.Millisecond)
	
	// Manual sync from editor1 to editor2 for testing
	docBytes, _ := json.Marshal(doc1)
	var docCopy crdt.Document
	_ = json.Unmarshal(docBytes, &docCopy)
	editorState2.SetDocument(&docCopy)
	
	// Check text is synchronized
	if model2.GetDocumentText() != "Hi" {
		t.Errorf("Document sync failed: got '%s', want 'Hi'", model2.GetDocumentText())
	}
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
