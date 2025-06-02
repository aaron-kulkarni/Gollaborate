package cursor

import (
	"gollaborate/crdt"
	"testing"
)

func TestGetCRDTPositionFromTextCoords(t *testing.T) {
	// Create a test document
	doc := &crdt.Document{
		Lines: []crdt.Line{
			{
				Characters: []crdt.Character{
					{Pos: []crdt.Identifier{{Digit: 1, Node: 1}}, Value: 'H'},
					{Pos: []crdt.Identifier{{Digit: 2, Node: 1}}, Value: 'e'},
					{Pos: []crdt.Identifier{{Digit: 3, Node: 1}}, Value: 'l'},
					{Pos: []crdt.Identifier{{Digit: 4, Node: 1}}, Value: 'l'},
					{Pos: []crdt.Identifier{{Digit: 5, Node: 1}}, Value: 'o'},
				},
			},
			{
				Characters: []crdt.Character{
					{Pos: []crdt.Identifier{{Digit: 6, Node: 1}}, Value: 'W'},
					{Pos: []crdt.Identifier{{Digit: 7, Node: 1}}, Value: 'o'},
					{Pos: []crdt.Identifier{{Digit: 8, Node: 1}}, Value: 'r'},
					{Pos: []crdt.Identifier{{Digit: 9, Node: 1}}, Value: 'l'},
					{Pos: []crdt.Identifier{{Digit: 10, Node: 1}}, Value: 'd'},
				},
			},
		},
	}

	manager := NewManager(doc, 1, "User 1", "#FF0000")

	// Test getting position of first character
	pos, err := manager.GetCRDTPositionFromTextCoords(1, 1)
	if err != nil {
		t.Fatalf("Failed to get CRDT position: %v", err)
	}
	if len(pos) != 1 || pos[0].Digit != 1 || pos[0].Node != 1 {
		t.Errorf("Expected position [{1 1}], got %v", pos)
	}

	// Test getting position of middle character
	pos, err = manager.GetCRDTPositionFromTextCoords(1, 3)
	if err != nil {
		t.Fatalf("Failed to get CRDT position: %v", err)
	}
	if len(pos) != 1 || pos[0].Digit != 3 || pos[0].Node != 1 {
		t.Errorf("Expected position [{3 1}], got %v", pos)
	}

	// Test getting position from second line
	pos, err = manager.GetCRDTPositionFromTextCoords(2, 1)
	if err != nil {
		t.Fatalf("Failed to get CRDT position: %v", err)
	}
	if len(pos) != 1 || pos[0].Digit != 6 || pos[0].Node != 1 {
		t.Errorf("Expected position [{6 1}], got %v", pos)
	}
}

func TestGetTextCoordsFromCRDTPosition(t *testing.T) {
	// Create a test document
	doc := &crdt.Document{
		Lines: []crdt.Line{
			{
				Characters: []crdt.Character{
					{Pos: []crdt.Identifier{{Digit: 1, Node: 1}}, Value: 'H'},
					{Pos: []crdt.Identifier{{Digit: 2, Node: 1}}, Value: 'e'},
					{Pos: []crdt.Identifier{{Digit: 3, Node: 1}}, Value: 'l'},
				},
			},
			{
				Characters: []crdt.Character{
					{Pos: []crdt.Identifier{{Digit: 4, Node: 1}}, Value: 'W'},
					{Pos: []crdt.Identifier{{Digit: 5, Node: 1}}, Value: 'o'},
				},
			},
		},
	}

	manager := NewManager(doc, 1, "User 1", "#FF0000")

	// Test converting CRDT position to text coordinates
	pos := []crdt.Identifier{{Digit: 1, Node: 1}}
	coords, err := manager.GetTextCoordsFromCRDTPosition(pos)
	if err != nil {
		t.Fatalf("Failed to get text coordinates: %v", err)
	}
	if coords.Line != 1 || coords.Column != 1 {
		t.Errorf("Expected coordinates (1, 1), got (%d, %d)", coords.Line, coords.Column)
	}

	// Test second line position
	pos = []crdt.Identifier{{Digit: 4, Node: 1}}
	coords, err = manager.GetTextCoordsFromCRDTPosition(pos)
	if err != nil {
		t.Fatalf("Failed to get text coordinates: %v", err)
	}
	if coords.Line != 2 || coords.Column != 1 {
		t.Errorf("Expected coordinates (2, 1), got (%d, %d)", coords.Line, coords.Column)
	}
}

func TestGetCRDTSelectionFromTextCoords(t *testing.T) {
	// Create a test document
	doc := &crdt.Document{
		Lines: []crdt.Line{
			{
				Characters: []crdt.Character{
					{Pos: []crdt.Identifier{{Digit: 1, Node: 1}}, Value: 'H'},
					{Pos: []crdt.Identifier{{Digit: 2, Node: 1}}, Value: 'e'},
					{Pos: []crdt.Identifier{{Digit: 3, Node: 1}}, Value: 'l'},
					{Pos: []crdt.Identifier{{Digit: 4, Node: 1}}, Value: 'l'},
					{Pos: []crdt.Identifier{{Digit: 5, Node: 1}}, Value: 'o'},
				},
			},
		},
	}

	manager := NewManager(doc, 1, "User 1", "#FF0000")

	// Test selection from (1,1) to (1,3)
	startPos, endPos, err := manager.GetCRDTSelectionFromTextCoords(1, 1, 1, 3)
	if err != nil {
		t.Fatalf("Failed to get CRDT selection: %v", err)
	}

	if len(startPos) != 1 || startPos[0].Digit != 1 {
		t.Errorf("Expected start position [{1 1}], got %v", startPos)
	}

	if len(endPos) != 1 || endPos[0].Digit != 3 {
		t.Errorf("Expected end position [{3 1}], got %v", endPos)
	}
}

func TestExtractTextFromSelection(t *testing.T) {
	// Create a test document with "Hello\nWorld"
	doc := &crdt.Document{
		Lines: []crdt.Line{
			{
				Characters: []crdt.Character{
					{Pos: []crdt.Identifier{{Digit: 1, Node: 1}}, Value: 'H'},
					{Pos: []crdt.Identifier{{Digit: 2, Node: 1}}, Value: 'e'},
					{Pos: []crdt.Identifier{{Digit: 3, Node: 1}}, Value: 'l'},
					{Pos: []crdt.Identifier{{Digit: 4, Node: 1}}, Value: 'l'},
					{Pos: []crdt.Identifier{{Digit: 5, Node: 1}}, Value: 'o'},
				},
			},
			{
				Characters: []crdt.Character{
					{Pos: []crdt.Identifier{{Digit: 6, Node: 1}}, Value: 'W'},
					{Pos: []crdt.Identifier{{Digit: 7, Node: 1}}, Value: 'o'},
					{Pos: []crdt.Identifier{{Digit: 8, Node: 1}}, Value: 'r'},
					{Pos: []crdt.Identifier{{Digit: 9, Node: 1}}, Value: 'l'},
					{Pos: []crdt.Identifier{{Digit: 10, Node: 1}}, Value: 'd'},
				},
			},
		},
	}

	manager := NewManager(doc, 1, "User 1", "#FF0000")

	// Test extracting "ell" from "Hello"
	startPos := []crdt.Identifier{{Digit: 2, Node: 1}} // 'e'
	endPos := []crdt.Identifier{{Digit: 4, Node: 1}}   // 'l' (exclusive)

	text, err := manager.ExtractTextFromSelection(startPos, endPos)
	if err != nil {
		t.Fatalf("Failed to extract text: %v", err)
	}

	expected := "el"
	if text != expected {
		t.Errorf("Expected text '%s', got '%s'", expected, text)
	}
}

func TestEmptyDocument(t *testing.T) {
	doc := &crdt.Document{Lines: []crdt.Line{}}
	manager := NewManager(doc, 1, "User 1", "#FF0000")

	// Test with empty document
	_, err := manager.GetCRDTPositionFromTextCoords(1, 1)
	if err == nil {
		t.Error("Expected error for empty document, got nil")
	}

	// Test empty position
	coords, err := manager.GetTextCoordsFromCRDTPosition([]crdt.Identifier{})
	if err != nil {
		t.Fatalf("Failed to get coords for empty position: %v", err)
	}
	if coords.Line != 1 || coords.Column != 1 {
		t.Errorf("Expected (1,1) for empty position, got (%d,%d)", coords.Line, coords.Column)
	}
}