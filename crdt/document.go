package crdt

import (
	"fmt"
	"sort"
	"strings"
)

// InsertCharacter inserts a character at the specified position in the document
func (d *Document) InsertCharacter(char rune, position []Identifier, clock int) error {
	if len(d.Lines) == 0 {
		d.Lines = append(d.Lines, Line{Characters: []Character{}})
	}

	newChar := Character{
		Pos:   position,
		Clock: clock,
		Value: char,
	}

	// Handle newline characters
	if char == '\n' {
		// Find the line where this character should be inserted
		lineIndex, charIndex := d.findInsertionPoint(position)
		
		// Split the line at the insertion point
		currentLine := d.Lines[lineIndex]
		
		// Create new line with characters after the insertion point
		newLine := Line{Characters: make([]Character, len(currentLine.Characters)-charIndex)}
		copy(newLine.Characters, currentLine.Characters[charIndex:])
		
		// Truncate current line and add newline character
		d.Lines[lineIndex].Characters = append(currentLine.Characters[:charIndex], newChar)
		
		// Insert the new line
		d.Lines = append(d.Lines[:lineIndex+1], append([]Line{newLine}, d.Lines[lineIndex+1:]...)...)
	} else {
		// Regular character insertion
		lineIndex, charIndex := d.findInsertionPoint(position)
		line := &d.Lines[lineIndex]
		
		// Insert character at the correct position
		line.Characters = append(line.Characters[:charIndex], append([]Character{newChar}, line.Characters[charIndex:]...)...)
	}

	return nil
}

// DeleteCharacter removes a character at the specified position
func (d *Document) DeleteCharacter(position []Identifier) error {
	lineIndex, charIndex, found := d.findCharacter(position)
	if !found {
		return fmt.Errorf("character not found at position")
	}

	char := d.Lines[lineIndex].Characters[charIndex]
	
	// Handle newline deletion
	if char.Value == '\n' {
		// Merge the next line with current line
		if lineIndex+1 < len(d.Lines) {
			// Remove the newline character
			d.Lines[lineIndex].Characters = append(d.Lines[lineIndex].Characters[:charIndex], d.Lines[lineIndex].Characters[charIndex+1:]...)
			
			// Merge next line's characters
			if lineIndex+1 < len(d.Lines) {
				d.Lines[lineIndex].Characters = append(d.Lines[lineIndex].Characters, d.Lines[lineIndex+1].Characters...)
				// Remove the merged line
				d.Lines = append(d.Lines[:lineIndex+1], d.Lines[lineIndex+2:]...)
			}
		} else {
			// Just remove the newline character if it's the last line
			d.Lines[lineIndex].Characters = append(d.Lines[lineIndex].Characters[:charIndex], d.Lines[lineIndex].Characters[charIndex+1:]...)
		}
	} else {
		// Regular character deletion
		line := &d.Lines[lineIndex]
		line.Characters = append(line.Characters[:charIndex], line.Characters[charIndex+1:]...)
	}

	return nil
}

// ToText converts the CRDT document to a plain text string
func (d *Document) ToText() string {
	var result strings.Builder
	
	for lineIndex, line := range d.Lines {
		for _, char := range line.Characters {
			if char.Value != '\n' {
				result.WriteRune(char.Value)
			}
		}
		
		// Add newline between lines (except for the last line)
		if lineIndex < len(d.Lines)-1 {
			result.WriteRune('\n')
		}
	}
	
	return result.String()
}

// FromText creates a CRDT document from a plain text string
func FromText(text string, nodeID int) *Document {
	doc := &Document{Lines: []Line{}}
	
	if text == "" {
		doc.Lines = append(doc.Lines, Line{Characters: []Character{}})
		return doc
	}
	
	lines := strings.Split(text, "\n")
	clock := 1
	
	for lineIndex, lineText := range lines {
		characters := make([]Character, 0, len(lineText))
		
		for _, char := range lineText {
			position := []Identifier{{Digit: clock, Node: nodeID}}
			characters = append(characters, Character{
				Pos:   position,
				Clock: clock,
				Value: char,
			})
			clock++
		}
		
		// Add newline character except for the last line
		if lineIndex < len(lines)-1 {
			position := []Identifier{{Digit: clock, Node: nodeID}}
			characters = append(characters, Character{
				Pos:   position,
				Clock: clock,
				Value: '\n',
			})
			clock++
		}
		
		doc.Lines = append(doc.Lines, Line{Characters: characters})
	}
	
	return doc
}

// GeneratePositionAt generates a position between two existing positions
func (d *Document) GeneratePositionAt(textLine, textColumn, nodeID int) ([]Identifier, error) {
	if len(d.Lines) == 0 {
		return []Identifier{{Digit: 1, Node: nodeID}}, nil
	}
	
	// Convert text coordinates to character index
	charIndex := 0
	for i := 0; i < textLine-1 && i < len(d.Lines); i++ {
		charIndex += len(d.Lines[i].Characters)
	}
	
	if textLine-1 < len(d.Lines) {
		charIndex += min(textColumn-1, len(d.Lines[textLine-1].Characters))
	}
	
	// Get all characters in document order
	allChars := d.getAllCharacters()
	
	// If no characters exist, return a simple position
	if len(allChars) == 0 {
		return []Identifier{{Digit: 1, Node: nodeID}}, nil
	}
	
	var prevPos, nextPos []Identifier
	
	if charIndex == 0 {
		// Insert at beginning
		nextPos = allChars[0].Pos
	} else if charIndex >= len(allChars) {
		// Insert at end
		prevPos = allChars[len(allChars)-1].Pos
	} else {
		// Insert between characters
		prevPos = allChars[charIndex-1].Pos
		nextPos = allChars[charIndex].Pos
	}
	
	return generatePositionBetween(prevPos, nextPos, nodeID), nil
}

// FindPositionAt finds the CRDT position at the given text coordinates
func (d *Document) FindPositionAt(textLine, textColumn int) ([]Identifier, error) {
	if textLine < 1 || textLine > len(d.Lines) {
		return nil, fmt.Errorf("line %d out of range", textLine)
	}
	
	line := d.Lines[textLine-1]
	if textColumn < 1 || textColumn > len(line.Characters)+1 {
		return nil, fmt.Errorf("column %d out of range", textColumn)
	}
	
	if textColumn <= len(line.Characters) {
		return line.Characters[textColumn-1].Pos, nil
	}
	
	// Position after last character
	if len(line.Characters) > 0 {
		return line.Characters[len(line.Characters)-1].Pos, nil
	}
	
	return []Identifier{}, nil
}

// findInsertionPoint finds where to insert a character with the given position
func (d *Document) findInsertionPoint(position []Identifier) (lineIndex, charIndex int) {
	allChars := d.getAllCharacters()
	
	// Find insertion point using position comparison
	for i, char := range allChars {
		if comparePositions(position, char.Pos) < 0 {
			return d.getLineAndCharIndex(i)
		}
	}
	
	// Insert at end
	if len(d.Lines) == 0 {
		return 0, 0
	}
	return len(d.Lines) - 1, len(d.Lines[len(d.Lines)-1].Characters)
}

// findCharacter finds a character with the given position
func (d *Document) findCharacter(position []Identifier) (lineIndex, charIndex int, found bool) {
	for lineIdx, line := range d.Lines {
		for charIdx, char := range line.Characters {
			if comparePositions(position, char.Pos) == 0 {
				return lineIdx, charIdx, true
			}
		}
	}
	return 0, 0, false
}

// getAllCharacters returns all characters in document order
func (d *Document) getAllCharacters() []Character {
	var allChars []Character
	for _, line := range d.Lines {
		allChars = append(allChars, line.Characters...)
	}
	
	// Sort by position
	sort.Slice(allChars, func(i, j int) bool {
		return comparePositions(allChars[i].Pos, allChars[j].Pos) < 0
	})
	
	return allChars
}

// getLineAndCharIndex converts a character index to line and character indices
func (d *Document) getLineAndCharIndex(charIndex int) (lineIndex, charIndexInLine int) {
	currentIndex := 0
	for lineIdx, line := range d.Lines {
		if currentIndex+len(line.Characters) > charIndex {
			return lineIdx, charIndex - currentIndex
		}
		currentIndex += len(line.Characters)
	}
	
	// If we get here, insert at the end
	if len(d.Lines) == 0 {
		return 0, 0
	}
	return len(d.Lines) - 1, len(d.Lines[len(d.Lines)-1].Characters)
}

// comparePositions compares two positions lexicographically
func comparePositions(pos1, pos2 []Identifier) int {
	minLen := min(len(pos1), len(pos2))
	
	for i := 0; i < minLen; i++ {
		if pos1[i].Digit != pos2[i].Digit {
			return pos1[i].Digit - pos2[i].Digit
		}
		if pos1[i].Node != pos2[i].Node {
			return pos1[i].Node - pos2[i].Node
		}
	}
	
	return len(pos1) - len(pos2)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}