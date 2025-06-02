package cursor

import (
	"fmt"
	"gollaborate/crdt"
	"strings"
)

// Manager handles cursor and selection tracking for collaborative editing
type Manager struct {
	document *crdt.Document
	userID   int
	userName string
	color    string
}

// NewManager creates a new cursor manager
func NewManager(document *crdt.Document, userID int, userName, color string) *Manager {
	return &Manager{
		document: document,
		userID:   userID,
		userName: userName,
		color:    color,
	}
}

// TextPosition represents a position in the GUI text (line, column)
type TextPosition struct {
	Line   int
	Column int
}

// GetCRDTPositionFromTextCoords converts GUI text coordinates to CRDT position
func (m *Manager) GetCRDTPositionFromTextCoords(line, column int) ([]crdt.Identifier, error) {
	if m.document == nil {
		return nil, fmt.Errorf("document is nil")
	}

	// Handle empty document
	if len(m.document.Lines) == 0 {
		return nil, fmt.Errorf("document is empty")
	}

	// Validate line number
	if line < 1 || line > len(m.document.Lines) {
		return nil, fmt.Errorf("line %d out of range (1-%d)", line, len(m.document.Lines))
	}

	lineIndex := line - 1
	documentLine := m.document.Lines[lineIndex]

	// Handle empty line
	if len(documentLine.Characters) == 0 {
		return []crdt.Identifier{}, nil
	}

	// Validate column number
	if column < 1 {
		return nil, fmt.Errorf("column must be >= 1")
	}

	// If column is beyond the line, return position after last character
	if column > len(documentLine.Characters) {
		lastChar := documentLine.Characters[len(documentLine.Characters)-1]
		return lastChar.Pos, nil
	}

	// Return the position of the character at the specified column
	columnIndex := column - 1
	return documentLine.Characters[columnIndex].Pos, nil
}

// GetTextCoordsFromCRDTPosition converts CRDT position to GUI text coordinates
func (m *Manager) GetTextCoordsFromCRDTPosition(position []crdt.Identifier) (TextPosition, error) {
	if m.document == nil {
		return TextPosition{}, fmt.Errorf("document is nil")
	}

	// Handle empty position (beginning of document)
	if len(position) == 0 {
		return TextPosition{Line: 1, Column: 1}, nil
	}

	// Search through all lines and characters to find the position
	for lineIndex, line := range m.document.Lines {
		for charIndex, char := range line.Characters {
			if identifiersEqual(char.Pos, position) {
				return TextPosition{
					Line:   lineIndex + 1,
					Column: charIndex + 1,
				}, nil
			}
		}
	}

	// If position not found, return end of document
	if len(m.document.Lines) == 0 {
		return TextPosition{Line: 1, Column: 1}, nil
	}

	lastLineIndex := len(m.document.Lines) - 1
	lastLine := m.document.Lines[lastLineIndex]
	return TextPosition{
		Line:   lastLineIndex + 1,
		Column: len(lastLine.Characters) + 1,
	}, nil
}

// GetCRDTSelectionFromTextCoords converts GUI selection to CRDT positions
func (m *Manager) GetCRDTSelectionFromTextCoords(startLine, startCol, endLine, endCol int) ([]crdt.Identifier, []crdt.Identifier, error) {
	startPos, err := m.GetCRDTPositionFromTextCoords(startLine, startCol)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get start position: %w", err)
	}

	endPos, err := m.GetCRDTPositionFromTextCoords(endLine, endCol)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get end position: %w", err)
	}

	return startPos, endPos, nil
}

// GetTextSelectionFromCRDTPositions converts CRDT selection to GUI coordinates
func (m *Manager) GetTextSelectionFromCRDTPositions(startPos, endPos []crdt.Identifier) (TextPosition, TextPosition, error) {
	start, err := m.GetTextCoordsFromCRDTPosition(startPos)
	if err != nil {
		return TextPosition{}, TextPosition{}, fmt.Errorf("failed to get start coordinates: %w", err)
	}

	end, err := m.GetTextCoordsFromCRDTPosition(endPos)
	if err != nil {
		return TextPosition{}, TextPosition{}, fmt.Errorf("failed to get end coordinates: %w", err)
	}

	return start, end, nil
}

// ExtractTextFromSelection returns the text content within a selection range
func (m *Manager) ExtractTextFromSelection(startPos, endPos []crdt.Identifier) (string, error) {
	startCoords, err := m.GetTextCoordsFromCRDTPosition(startPos)
	if err != nil {
		return "", err
	}

	endCoords, err := m.GetTextCoordsFromCRDTPosition(endPos)
	if err != nil {
		return "", err
	}

	return m.extractTextBetweenCoords(startCoords, endCoords), nil
}

// extractTextBetweenCoords extracts text between two text coordinates
func (m *Manager) extractTextBetweenCoords(start, end TextPosition) string {
	if m.document == nil || len(m.document.Lines) == 0 {
		return ""
	}

	var result strings.Builder

	// Ensure start comes before end
	if start.Line > end.Line || (start.Line == end.Line && start.Column > end.Column) {
		start, end = end, start
	}

	for lineNum := start.Line; lineNum <= end.Line && lineNum <= len(m.document.Lines); lineNum++ {
		lineIndex := lineNum - 1
		line := m.document.Lines[lineIndex]

		startCol := 1
		endCol := len(line.Characters)

		if lineNum == start.Line {
			startCol = start.Column
		}
		if lineNum == end.Line {
			endCol = end.Column - 1 // Exclusive end
		}

		for col := startCol; col <= endCol && col <= len(line.Characters); col++ {
			charIndex := col - 1
			result.WriteRune(line.Characters[charIndex].Value)
		}

		// Add newline if not the last line in selection
		if lineNum < end.Line {
			result.WriteRune('\n')
		}
	}

	return result.String()
}

// identifiersEqual compares two identifier slices for equality
func identifiersEqual(a, b []crdt.Identifier) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Digit != b[i].Digit || a[i].Node != b[i].Node {
			return false
		}
	}

	return true
}

// UpdateDocument updates the document reference in the cursor manager
func (m *Manager) UpdateDocument(document *crdt.Document) {
	m.document = document
}

// GetUserInfo returns the user information for this cursor manager
func (m *Manager) GetUserInfo() (int, string, string) {
	return m.userID, m.userName, m.color
}