package messages

import (
	"gollaborate/crdt"
	"testing"
)

func TestMessageSerialization(t *testing.T) {
	// Test Operation message
	position := []crdt.Identifier{
		{Digit: 10, Node: 1},
		{Digit: 20, Node: 2},
	}
	
	op := NewInsertOperation(position, 'A', 1, 5)
	msg := NewOperationMessage(op)
	
	// Serialize
	data, err := msg.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize message: %v", err)
	}
	
	// Deserialize
	deserializedMsg, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize message: %v", err)
	}
	
	// Verify
	if deserializedMsg.Type != MessageTypeOperation {
		t.Errorf("Expected type %s, got %s", MessageTypeOperation, deserializedMsg.Type)
	}
	
	if deserializedMsg.Operation.Type != OperationTypeInsert {
		t.Errorf("Expected operation type %s, got %s", OperationTypeInsert, deserializedMsg.Operation.Type)
	}
	
	if deserializedMsg.Operation.Character != 'A' {
		t.Errorf("Expected character 'A', got '%c'", deserializedMsg.Operation.Character)
	}
	
	if deserializedMsg.Operation.UserID != 1 {
		t.Errorf("Expected user ID 1, got %d", deserializedMsg.Operation.UserID)
	}
	
	if len(deserializedMsg.Operation.Position) != 2 {
		t.Errorf("Expected position length 2, got %d", len(deserializedMsg.Operation.Position))
	}
}

func TestDocumentMessage(t *testing.T) {
	// Create a simple document
	doc := &crdt.Document{
		Lines: []crdt.Line{
			{
				Characters: []crdt.Character{
					{
						Pos:   []crdt.Identifier{{Digit: 1, Node: 1}},
						Clock: 1,
						Value: 'H',
					},
					{
						Pos:   []crdt.Identifier{{Digit: 2, Node: 1}},
						Clock: 2,
						Value: 'i',
					},
				},
			},
		},
	}
	
	msg := NewSyncMessage(doc, 1)
	
	// Serialize
	data, err := msg.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize document message: %v", err)
	}
	
	// Deserialize
	deserializedMsg, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize document message: %v", err)
	}
	
	// Verify
	if deserializedMsg.Type != MessageTypeSync {
		t.Errorf("Expected type %s, got %s", MessageTypeSync, deserializedMsg.Type)
	}
	
	if len(deserializedMsg.Document.Lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(deserializedMsg.Document.Lines))
	}
	
	if len(deserializedMsg.Document.Lines[0].Characters) != 2 {
		t.Errorf("Expected 2 characters, got %d", len(deserializedMsg.Document.Lines[0].Characters))
	}
	
	if deserializedMsg.Document.Lines[0].Characters[0].Value != 'H' {
		t.Errorf("Expected first character 'H', got '%c'", deserializedMsg.Document.Lines[0].Characters[0].Value)
	}
}

func TestCursorMessage(t *testing.T) {
	position := []crdt.Identifier{
		{Digit: 5, Node: 2},
		{Digit: 10, Node: 2},
	}
	
	msg := NewCursorMessage(position, 2, "Alice", "#00FF00")
	
	// Serialize
	data, err := msg.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize cursor message: %v", err)
	}
	
	// Deserialize
	deserializedMsg, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize cursor message: %v", err)
	}
	
	// Verify
	if deserializedMsg.Type != MessageTypeCursor {
		t.Errorf("Expected type %s, got %s", MessageTypeCursor, deserializedMsg.Type)
	}
	
	if deserializedMsg.Cursor.UserID != 2 {
		t.Errorf("Expected user ID 2, got %d", deserializedMsg.Cursor.UserID)
	}
	
	if deserializedMsg.Cursor.UserName != "Alice" {
		t.Errorf("Expected user name 'Alice', got '%s'", deserializedMsg.Cursor.UserName)
	}
	
	if deserializedMsg.Cursor.Color != "#00FF00" {
		t.Errorf("Expected color '#00FF00', got '%s'", deserializedMsg.Cursor.Color)
	}
	
	if len(deserializedMsg.Cursor.Position) != 2 {
		t.Errorf("Expected position length 2, got %d", len(deserializedMsg.Cursor.Position))
	}
	
	if deserializedMsg.Cursor.Position[0].Digit != 5 || deserializedMsg.Cursor.Position[0].Node != 2 {
		t.Errorf("Expected first position {5 2}, got {%d %d}", 
			deserializedMsg.Cursor.Position[0].Digit, deserializedMsg.Cursor.Position[0].Node)
	}
}

func TestSelectionMessage(t *testing.T) {
	startPos := []crdt.Identifier{{Digit: 1, Node: 1}}
	endPos := []crdt.Identifier{{Digit: 5, Node: 1}}
	
	msg := NewSelectionMessage(startPos, endPos, 3, "Bob", "#0000FF")
	
	// Serialize
	data, err := msg.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize selection message: %v", err)
	}
	
	// Deserialize
	deserializedMsg, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize selection message: %v", err)
	}
	
	// Verify
	if deserializedMsg.Type != MessageTypeSelection {
		t.Errorf("Expected type %s, got %s", MessageTypeSelection, deserializedMsg.Type)
	}
	
	if deserializedMsg.Selection.UserID != 3 {
		t.Errorf("Expected user ID 3, got %d", deserializedMsg.Selection.UserID)
	}
	
	if deserializedMsg.Selection.UserName != "Bob" {
		t.Errorf("Expected user name 'Bob', got '%s'", deserializedMsg.Selection.UserName)
	}
	
	if deserializedMsg.Selection.Color != "#0000FF" {
		t.Errorf("Expected color '#0000FF', got '%s'", deserializedMsg.Selection.Color)
	}
	
	if len(deserializedMsg.Selection.StartPosition) != 1 {
		t.Errorf("Expected start position length 1, got %d", len(deserializedMsg.Selection.StartPosition))
	}
	
	if len(deserializedMsg.Selection.EndPosition) != 1 {
		t.Errorf("Expected end position length 1, got %d", len(deserializedMsg.Selection.EndPosition))
	}
	
	if deserializedMsg.Selection.StartPosition[0].Digit != 1 {
		t.Errorf("Expected start position digit 1, got %d", deserializedMsg.Selection.StartPosition[0].Digit)
	}
	
	if deserializedMsg.Selection.EndPosition[0].Digit != 5 {
		t.Errorf("Expected end position digit 5, got %d", deserializedMsg.Selection.EndPosition[0].Digit)
	}
}

func TestClearSelectionMessage(t *testing.T) {
	msg := NewSelectionMessage(nil, nil, 4, "Carol", "#FF00FF")
	
	// Serialize
	data, err := msg.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize clear selection message: %v", err)
	}
	
	// Deserialize
	deserializedMsg, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize clear selection message: %v", err)
	}
	
	// Verify
	if deserializedMsg.Type != MessageTypeSelection {
		t.Errorf("Expected type %s, got %s", MessageTypeSelection, deserializedMsg.Type)
	}
	
	if deserializedMsg.Selection.StartPosition != nil {
		t.Errorf("Expected nil start position for clear selection, got %v", deserializedMsg.Selection.StartPosition)
	}
	
	if deserializedMsg.Selection.EndPosition != nil {
		t.Errorf("Expected nil end position for clear selection, got %v", deserializedMsg.Selection.EndPosition)
	}
	
	if deserializedMsg.Selection.UserID != 4 {
		t.Errorf("Expected user ID 4, got %d", deserializedMsg.Selection.UserID)
	}
}