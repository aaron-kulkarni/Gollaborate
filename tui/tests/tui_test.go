package tests

import (
	"testing"

	"gollaborate/crdt"
	"gollaborate/shared"
	"gollaborate/tui/core"
)

// TestTUIInitialization verifies that the TUI initializes properly
func TestTUIInitialization(t *testing.T) {
	// Create a document and editor state
	doc := crdt.FromText("Hello, world!", 1)
	editorState := shared.NewEditorState(doc, 1)

	// Mock TUI initialization
	m := core.InitializeModelForTesting(editorState, 1, "blue")
	if m == nil {
		t.Fatal("Failed to initialize TUI model")
	}

	// Verify document content
	text := m.GetDocumentText()
	if text != "Hello, world!" {
		t.Errorf("Document text incorrect: got '%s', want 'Hello, world!'", text)
	}
}

// TestTUIKeyboardHandling verifies that keyboard input works correctly
func TestTUIKeyboardHandling(t *testing.T) {
	// Create a document and editor state
	doc := crdt.FromText("", 1)
	editorState := shared.NewEditorState(doc, 1)

	// Mock TUI initialization
	m := core.InitializeModelForTesting(editorState, 1, "blue")
	if m == nil {
		t.Fatal("Failed to initialize TUI model")
	}

	// Simulate typing "abc"
	m.SimulateKeyPress("a")
	m.SimulateKeyPress("b")
	m.SimulateKeyPress("c")
	
	// Verify document content
	text := m.GetDocumentText()
	if text != "abc" {
		t.Errorf("Document text after typing incorrect: got '%s', want 'abc'", text)
	}
	
	// Test cursor position
	cursorX, cursorY := m.GetCursorPosition()
	if cursorX != 4 || cursorY != 1 { // Cursor should be at position (4,1) after typing 3 chars
		t.Errorf("Cursor position incorrect: got (%d,%d), want (4,1)", cursorX, cursorY)
	}
}

// TestTUICRDTIntegration verifies that CRDT operations are correctly applied
func TestTUICRDTIntegration(t *testing.T) {
	// Create a document and editor state
	doc := crdt.FromText("abc", 1)
	editorState := shared.NewEditorState(doc, 1)

	// Mock TUI initialization
	m := core.InitializeModelForTesting(editorState, 1, "blue")
	if m == nil {
		t.Fatal("Failed to initialize TUI model")
	}

	// Position cursor and delete a character
	// Test cursor position and delete a character
	m.SetCursorPosition(2, 1) // Position at 'b'
	m.SimulateKeyPress("backspace")
	
	// Verify document content
	text := m.GetDocumentText()
	if text != "bc" {
		t.Errorf("Document text after deletion incorrect: got '%s', want 'bc'", text)
	}
	
	// Insert at current position (which should now be position 1 after the deletion)
	m.SimulateKeyPress("x")
	text = m.GetDocumentText()
	if text != "xbc" {
		t.Errorf("Document text after insertion incorrect: got '%s', want 'xbc'", text)
	}
}

// TestTUIMultilineEditing verifies that multiline editing works correctly
func TestTUIMultilineEditing(t *testing.T) {
	// Create a document and editor state
	doc := crdt.FromText("Line 1", 1)
	editorState := shared.NewEditorState(doc, 1)

	// Mock TUI initialization
	m := core.InitializeModelForTesting(editorState, 1, "blue")
	if m == nil {
		t.Fatal("Failed to initialize TUI model")
	}

	// Position cursor at end of line and press enter
	m.SetCursorPosition(7, 1) // End of "Line 1"
	m.SimulateKeyPress("enter")
	
	// Type on the new line
	m.SimulateKeyPress("L")
	m.SimulateKeyPress("i")
	m.SimulateKeyPress("n")
	m.SimulateKeyPress("e")
	m.SimulateKeyPress(" ")
	m.SimulateKeyPress("2")
	
	// Verify document content
	text := m.GetDocumentText()
	expected := "Line 1\nLine 2"
	if text != expected {
		t.Errorf("Multiline text incorrect: got '%s', want '%s'", text, expected)
	}
}