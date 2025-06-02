package gui

import (
	"testing"

	"gollaborate/crdt"
	"gollaborate/messages"
)

func TestEditorState(t *testing.T) {
	// Test creating new editor state
	editorState := NewEditorState(nil, 1)
	
	if editorState.nodeID != 1 {
		t.Errorf("Expected nodeID 1, got %d", editorState.nodeID)
	}
	
	if editorState.clock != 1 {
		t.Errorf("Expected initial clock 1, got %d", editorState.clock)
	}
	
	if editorState.document == nil {
		t.Error("Expected document to be initialized")
	}
	
	if editorState.cursorMgr == nil {
		t.Error("Expected cursor manager to be initialized")
	}
}

func TestNextClock(t *testing.T) {
	editorState := NewEditorState(nil, 1)
	
	first := editorState.nextClock()
	second := editorState.nextClock()
	
	if first != 2 {
		t.Errorf("Expected first clock 2, got %d", first)
	}
	
	if second != 3 {
		t.Errorf("Expected second clock 3, got %d", second)
	}
}

func TestApplyOperation(t *testing.T) {
	editorState := NewEditorState(nil, 1)
	
	// Test insert operation
	position := []crdt.Identifier{{Digit: 1, Node: 1}}
	insertOp := messages.NewInsertOperation(position, 'A', 1, 1)
	
	err := editorState.applyOperation(insertOp)
	if err != nil {
		t.Fatalf("Failed to apply insert operation: %v", err)
	}
	
	text := editorState.document.ToText()
	if text != "A" {
		t.Errorf("Expected text 'A', got '%s'", text)
	}
	
	// Test delete operation
	deleteOp := messages.NewDeleteOperation(position, 1, 2)
	err = editorState.applyOperation(deleteOp)
	if err != nil {
		t.Fatalf("Failed to apply delete operation: %v", err)
	}
	
	text = editorState.document.ToText()
	if text != "" {
		t.Errorf("Expected empty text after deletion, got '%s'", text)
	}
}

func TestDetectChanges(t *testing.T) {
	editorState := NewEditorState(nil, 1)
	
	// Test insertion
	operations := editorState.detectChanges("", "A")
	if len(operations) != 1 {
		t.Errorf("Expected 1 operation for insertion, got %d", len(operations))
	}
	if operations[0].Type != messages.OperationTypeInsert {
		t.Errorf("Expected insert operation, got %s", operations[0].Type)
	}
	if operations[0].Character != 'A' {
		t.Errorf("Expected character 'A', got '%c'", operations[0].Character)
	}
	
	// Test deletion
	operations = editorState.detectChanges("A", "")
	if len(operations) != 1 {
		t.Errorf("Expected 1 operation for deletion, got %d", len(operations))
	}
	if operations[0].Type != messages.OperationTypeDelete {
		t.Errorf("Expected delete operation, got %s", operations[0].Type)
	}
	
	// Test multiple insertions
	operations = editorState.detectChanges("", "ABC")
	if len(operations) != 3 {
		t.Errorf("Expected 3 operations for 'ABC' insertion, got %d", len(operations))
	}
}

func TestProcessTextChange(t *testing.T) {
	editorState := NewEditorState(nil, 1)
	editorState.lastText = ""
	
	// Simulate text change
	editorState.processTextChange("Hello")
	
	// Check that document was updated
	text := editorState.document.ToText()
	if text != "Hello" {
		t.Errorf("Expected document text 'Hello', got '%s'", text)
	}
	
	if editorState.lastText != "Hello" {
		t.Errorf("Expected lastText 'Hello', got '%s'", editorState.lastText)
	}
}

func TestCRDTConsistency(t *testing.T) {
	editorState := NewEditorState(nil, 1)
	
	// Apply a series of operations
	testTexts := []string{"H", "He", "Hel", "Hell", "Hello"}
	
	for _, text := range testTexts {
		editorState.processTextChange(text)
		
		// Verify document consistency
		docText := editorState.document.ToText()
		if docText != text {
			t.Errorf("Document inconsistency: expected '%s', got '%s'", text, docText)
		}
	}
	
	// Test deletion
	editorState.processTextChange("Hell")
	docText := editorState.document.ToText()
	if docText != "Hell" {
		t.Errorf("Expected 'Hell' after deletion, got '%s'", docText)
	}
}

func TestMultilineOperations(t *testing.T) {
	editorState := NewEditorState(nil, 1)
	
	// Test multiline text
	multilineText := "Line1\nLine2\nLine3"
	editorState.processTextChange(multilineText)
	
	docText := editorState.document.ToText()
	if docText != multilineText {
		t.Errorf("Expected multiline text '%s', got '%s'", multilineText, docText)
	}
	
	// Verify document structure
	if len(editorState.document.Lines) != 3 {
		t.Errorf("Expected 3 lines in document, got %d", len(editorState.document.Lines))
	}
}