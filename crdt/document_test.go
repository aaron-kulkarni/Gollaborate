package crdt

import (
	"testing"
)

func TestFromText(t *testing.T) {
	// Test empty text
	doc := FromText("", 1)
	if len(doc.Lines) != 1 {
		t.Errorf("Expected 1 line for empty text, got %d", len(doc.Lines))
	}
	if len(doc.Lines[0].Characters) != 0 {
		t.Errorf("Expected 0 characters for empty text, got %d", len(doc.Lines[0].Characters))
	}

	// Test single line
	doc = FromText("Hello", 1)
	if len(doc.Lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(doc.Lines))
	}
	if len(doc.Lines[0].Characters) != 5 {
		t.Errorf("Expected 5 characters, got %d", len(doc.Lines[0].Characters))
	}
	if doc.Lines[0].Characters[0].Value != 'H' {
		t.Errorf("Expected first character 'H', got '%c'", doc.Lines[0].Characters[0].Value)
	}

	// Test multiple lines
	doc = FromText("Hello\nWorld", 1)
	if len(doc.Lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(doc.Lines))
	}
	if len(doc.Lines[0].Characters) != 6 { // 5 chars + newline
		t.Errorf("Expected 6 characters in first line, got %d", len(doc.Lines[0].Characters))
	}
	if doc.Lines[0].Characters[5].Value != '\n' {
		t.Errorf("Expected newline character, got '%c'", doc.Lines[0].Characters[5].Value)
	}
}

func TestToText(t *testing.T) {
	// Test empty document
	doc := &Document{Lines: []Line{{Characters: []Character{}}}}
	text := doc.ToText()
	if text != "" {
		t.Errorf("Expected empty text, got '%s'", text)
	}

	// Test single line
	doc = FromText("Hello", 1)
	text = doc.ToText()
	if text != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", text)
	}

	// Test multiple lines
	doc = FromText("Hello\nWorld", 1)
	text = doc.ToText()
	if text != "Hello\nWorld" {
		t.Errorf("Expected 'Hello\\nWorld', got '%s'", text)
	}
}

func TestInsertCharacter(t *testing.T) {
	doc := FromText("Hello", 1)
	
	// Insert at beginning
	position := []Identifier{{Digit: 0, Node: 1}}
	err := doc.InsertCharacter('X', position, 10)
	if err != nil {
		t.Fatalf("Failed to insert character: %v", err)
	}
	
	text := doc.ToText()
	if text != "XHello" {
		t.Errorf("Expected 'XHello', got '%s'", text)
	}

	// Insert in middle
	doc = FromText("Hello", 1)
	position = []Identifier{{Digit: 3, Node: 1}}
	err = doc.InsertCharacter('X', position, 10)
	if err != nil {
		t.Fatalf("Failed to insert character: %v", err)
	}
	
	// Result depends on position ordering, but should contain X
	text = doc.ToText()
	if len(text) != 6 {
		t.Errorf("Expected length 6, got %d", len(text))
	}
}

func TestInsertNewline(t *testing.T) {
	doc := FromText("Hello", 1)
	
	// Insert newline in middle
	position := []Identifier{{Digit: 3, Node: 1}}
	err := doc.InsertCharacter('\n', position, 10)
	if err != nil {
		t.Fatalf("Failed to insert newline: %v", err)
	}
	
	if len(doc.Lines) < 2 {
		t.Errorf("Expected at least 2 lines after inserting newline, got %d", len(doc.Lines))
	}
}

func TestDeleteCharacter(t *testing.T) {
	doc := FromText("Hello", 1)
	
	// Get position of first character
	if len(doc.Lines) == 0 || len(doc.Lines[0].Characters) == 0 {
		t.Fatal("Document should have characters")
	}
	
	position := doc.Lines[0].Characters[0].Pos
	err := doc.DeleteCharacter(position)
	if err != nil {
		t.Fatalf("Failed to delete character: %v", err)
	}
	
	text := doc.ToText()
	if text != "ello" {
		t.Errorf("Expected 'ello', got '%s'", text)
	}
}

func TestDeleteNonExistentCharacter(t *testing.T) {
	doc := FromText("Hello", 1)
	
	// Try to delete character that doesn't exist
	position := []Identifier{{Digit: 999, Node: 999}}
	err := doc.DeleteCharacter(position)
	if err == nil {
		t.Error("Expected error when deleting non-existent character")
	}
}

func TestGeneratePositionAt(t *testing.T) {
	doc := FromText("Hello", 1)
	
	// Generate position at beginning
	position, err := doc.GeneratePositionAt(1, 1, 2)
	if err != nil {
		t.Fatalf("Failed to generate position: %v", err)
	}
	if len(position) == 0 {
		t.Error("Expected non-empty position")
	}
	
	// Generate position at end
	position, err = doc.GeneratePositionAt(1, 6, 2)
	if err != nil {
		t.Fatalf("Failed to generate position at end: %v", err)
	}
	if len(position) == 0 {
		t.Error("Expected non-empty position")
	}
}

func TestFindPositionAt(t *testing.T) {
	doc := FromText("Hello", 1)
	
	// Find position at beginning
	position, err := doc.FindPositionAt(1, 1)
	if err != nil {
		t.Fatalf("Failed to find position: %v", err)
	}
	if len(position) == 0 {
		t.Error("Expected non-empty position")
	}
	
	// Find position beyond line
	_, err = doc.FindPositionAt(1, 100)
	if err == nil {
		t.Error("Expected error for position beyond line")
	}
	
	// Find position on non-existent line
	_, err = doc.FindPositionAt(100, 1)
	if err == nil {
		t.Error("Expected error for non-existent line")
	}
}

func TestComparePositions(t *testing.T) {
	pos1 := []Identifier{{Digit: 1, Node: 1}}
	pos2 := []Identifier{{Digit: 2, Node: 1}}
	pos3 := []Identifier{{Digit: 1, Node: 2}}
	
	// Test digit comparison
	if comparePositions(pos1, pos2) >= 0 {
		t.Error("pos1 should be less than pos2")
	}
	if comparePositions(pos2, pos1) <= 0 {
		t.Error("pos2 should be greater than pos1")
	}
	
	// Test node comparison
	if comparePositions(pos1, pos3) >= 0 {
		t.Error("pos1 should be less than pos3")
	}
	
	// Test equality
	pos4 := []Identifier{{Digit: 1, Node: 1}}
	if comparePositions(pos1, pos4) != 0 {
		t.Error("pos1 should equal pos4")
	}
	
	// Test different lengths
	posLong := []Identifier{{Digit: 1, Node: 1}, {Digit: 1, Node: 1}}
	if comparePositions(pos1, posLong) >= 0 {
		t.Error("shorter position should be less than longer position with same prefix")
	}
}

func TestRoundTripTextConversion(t *testing.T) {
	testTexts := []string{
		"",
		"Hello",
		"Hello\nWorld",
		"Line1\nLine2\nLine3",
		"Single line with spaces",
		"Multiple\n\nEmpty\n\nLines",
	}
	
	for _, originalText := range testTexts {
		doc := FromText(originalText, 1)
		convertedText := doc.ToText()
		
		if convertedText != originalText {
			t.Errorf("Round-trip failed for text '%s': got '%s'", originalText, convertedText)
		}
	}
}

func TestComplexOperations(t *testing.T) {
	doc := FromText("Hello World", 1)
	
	// Insert multiple characters
	pos1, _ := doc.GeneratePositionAt(1, 6, 2)
	doc.InsertCharacter(',', pos1, 10)
	
	pos2, _ := doc.GeneratePositionAt(1, 7, 2)
	doc.InsertCharacter(' ', pos2, 11)
	
	text := doc.ToText()
	if len(text) != 13 { // Original 11 + 2 inserted
		t.Errorf("Expected length 13 after insertions, got %d", len(text))
	}
	
	// Delete a character
	if len(doc.Lines) > 0 && len(doc.Lines[0].Characters) > 0 {
		firstCharPos := doc.Lines[0].Characters[0].Pos
		doc.DeleteCharacter(firstCharPos)
		
		newText := doc.ToText()
		if len(newText) != 12 {
			t.Errorf("Expected length 12 after deletion, got %d", len(newText))
		}
	}
}