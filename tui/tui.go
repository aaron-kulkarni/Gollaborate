package core

import (
	"fmt"
	"sync"

	"gollaborate/crdt"
	"gollaborate/messages"
	"gollaborate/shared"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	doc         *crdt.Document
	cursorX     int // column (1-based)
	cursorY     int // line (1-based)
	status      string
	editorState *shared.EditorState
	userID      int
	userColor   string
	userName    string
	clock       int
	program     *tea.Program
	mutex       sync.Mutex

	// Selection state
	selectionActive bool
	selStartX       int
	selStartY       int
}

func initialModel(editorState *shared.EditorState, userID int, userColor string) *model {
	// Use the document from the editor state
	doc := editorState.Document()
	return &model{
		doc:         doc,
		cursorX:     1,
		cursorY:     1,
		status:      "Ready",
		editorState: editorState,
		userID:      userID,
		userColor:   userColor,
		userName:    fmt.Sprintf("User-%d", userID),
		clock:       1,
		mutex:       sync.Mutex{},
		selectionActive: false,
		selStartX:       0,
		selStartY:       0,
	}
}

func (m *model) Init() tea.Cmd {
	// Start message receiver in the background
	go m.listenForMessages()
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "ctrl+s":
			m.status = "Saved"
		case "backspace", "delete":
			if m.selectionActive {
				m.deleteSelection()
				m.selectionActive = false
				m.sendCursorUpdate()
			} else {
				// Delete character before cursor
				if m.cursorX > 1 {
					pos, err := m.doc.FindPositionAt(m.cursorY, m.cursorX-1)
					if err == nil {
						_ = m.doc.DeleteCharacter(pos)
						// Send delete operation to peers
						m.sendDeleteOperation(pos)
						m.cursorX--
						m.sendCursorUpdate()
					}
				} else if m.cursorY > 1 {
					// Handle backspace at start of line (merge lines)
					prevLineLen := len(m.doc.Lines[m.cursorY-2].Characters)
					pos, err := m.doc.FindPositionAt(m.cursorY-1, prevLineLen+1)
					if err == nil {
						_ = m.doc.DeleteCharacter(pos)
						// Send delete operation to peers
						m.sendDeleteOperation(pos)
						m.cursorY--
						m.cursorX = prevLineLen + 1
						m.sendCursorUpdate()
					}
				}
			}
		case "shift+left":
			// Start or extend selection to the left
			if !m.selectionActive {
				m.selectionActive = true
				m.selStartX = m.cursorX
				m.selStartY = m.cursorY
			}
			if m.cursorX > 1 {
				m.cursorX--
			}
		case "shift+right":
			// Start or extend selection to the right
			if !m.selectionActive {
				m.selectionActive = true
				m.selStartX = m.cursorX
				m.selStartY = m.cursorY
			}
			lineLen := 0
			if m.cursorY-1 < len(m.doc.Lines) {
				lineLen = len(m.doc.Lines[m.cursorY-1].Characters)
			}
			if m.cursorX <= lineLen {
				m.cursorX++
			}
		case "shift+up":
			if !m.selectionActive {
				m.selectionActive = true
				m.selStartX = m.cursorX
				m.selStartY = m.cursorY
			}
			if m.cursorY > 1 {
				m.cursorY--
				lineLen := len(m.doc.Lines[m.cursorY-1].Characters)
				if m.cursorX > lineLen+1 {
					m.cursorX = lineLen + 1
				}
			}
		case "shift+down":
			if !m.selectionActive {
				m.selectionActive = true
				m.selStartX = m.cursorX
				m.selStartY = m.cursorY
			}
			if m.cursorY < len(m.doc.Lines) {
				m.cursorY++
				lineLen := len(m.doc.Lines[m.cursorY-1].Characters)
				if m.cursorX > lineLen+1 {
					m.cursorX = lineLen + 1
				}
			}
		case "esc":
			// Clear selection
			m.selectionActive = false
		case "left":
			// Handle cursor movement
			if m.cursorX > 1 {
				m.cursorX--
			}
			m.selectionActive = false
		case "right":
			lineLen := 0
			if m.cursorY-1 < len(m.doc.Lines) {
				lineLen = len(m.doc.Lines[m.cursorY-1].Characters)
			}
			if m.cursorX <= lineLen {
				m.cursorX++
			}
			m.selectionActive = false
		case "up":
			if m.cursorY > 1 {
				m.cursorY--
				lineLen := len(m.doc.Lines[m.cursorY-1].Characters)
				if m.cursorX > lineLen+1 {
					m.cursorX = lineLen + 1
				}
			}
			m.selectionActive = false
		case "down":
			if m.cursorY < len(m.doc.Lines) {
				m.cursorY++
				lineLen := len(m.doc.Lines[m.cursorY-1].Characters)
				if m.cursorX > lineLen+1 {
					m.cursorX = lineLen + 1
				}
			}
			m.selectionActive = false

		// (handled above, moved for selection support)
		case "enter":
			pos, err := m.doc.GeneratePositionAt(m.cursorY, m.cursorX, m.userID)
			if err == nil {
				m.clock++
				_ = m.doc.InsertCharacter('\n', pos, m.clock)
				// Send insert operation to peers
				m.sendInsertOperation(pos, '\n')
				m.cursorY++
				m.cursorX = 1
				m.sendCursorUpdate()
			}
		default:
			// Insert printable characters
			r := []rune(msg.String())
			if len(r) == 1 && r[0] >= 32 && r[0] != 127 {
				if m.selectionActive {
					// Replace selection with character
					m.deleteSelection()
					pos, err := m.doc.GeneratePositionAt(m.cursorY, m.cursorX, m.userID)
					if err == nil {
						m.clock++
						_ = m.doc.InsertCharacter(r[0], pos, m.clock)
						m.sendInsertOperation(pos, r[0])
						m.cursorX++
						m.sendCursorUpdate()
					}
					m.selectionActive = false
				} else {
					pos, err := m.doc.GeneratePositionAt(m.cursorY, m.cursorX, m.userID)
					if err == nil {
						m.clock++
						_ = m.doc.InsertCharacter(r[0], pos, m.clock)
						// Send insert operation to peers
						m.sendInsertOperation(pos, r[0])
						m.cursorX++
						m.sendCursorUpdate()
					}
				}
			}
		}
	case networkMessageUpdate:
		// Handle incoming network messages
		m.handleMessage(msg.message)
		// Bubbletea doesn't support Message type as a message, so using our custom handler instead
	}
	return m, nil
}

func (m *model) sendCursorUpdate() {
	// Convert cursor position to CRDT position
	pos, err := m.doc.FindPositionAt(m.cursorY, m.cursorX)
	if err != nil {
		return
	}

	connections := m.editorState.Connections()
	for _, conn := range connections {
		_ = messages.SendCursor(conn, pos, m.userID, m.userName, m.userColor)
	}
}

func (m *model) sendInsertOperation(pos []crdt.Identifier, char rune) {
	operation := messages.NewInsertOperation(pos, char, m.userID, m.clock)
	connections := m.editorState.Connections()
	for _, conn := range connections {
		_ = messages.SendOperation(conn, operation)
	}
}

func (m *model) sendDeleteOperation(pos []crdt.Identifier) {
	operation := messages.NewDeleteOperation(pos, m.userID, m.clock)
	connections := m.editorState.Connections()
	for _, conn := range connections {
		_ = messages.SendOperation(conn, operation)
	}
}

// networkMessageUpdate is a custom message type for tea.Msg
type networkMessageUpdate struct {
	message *messages.Message
}

// listenForMessages listens for incoming messages from peers in a background goroutine
func (m *model) listenForMessages() {
	// Register as a message listener to the editor state
	m.editorState.AddMessageListener(func(msg *messages.Message) {
		// When a message is received, send it to the TUI update loop via a tea.Cmd
		if m.program != nil {
			m.program.Send(networkMessageUpdate{message: msg})
		}
	})
}

func (m *model) handleMessage(msg *messages.Message) {
	switch msg.Type {
	case messages.MessageTypeCursor:
		if msg.Cursor.UserID != m.userID {
			// Convert CRDT position to text coordinates
			// This would need to be implemented
			m.status = fmt.Sprintf("Cursor moved by %s", msg.Cursor.UserName)
		}
	case messages.MessageTypeSelection:
		if msg.Selection.UserID != m.userID {
			m.status = fmt.Sprintf("Selection updated by %s", msg.Selection.UserName)
			// Handle selection logic here
		}
	case messages.MessageTypeOperation:
		if msg.Operation.UserID != m.userID {
			op := msg.Operation
			// Do NOT apply the operation to the document here!
			// The EditorState already did it.
			switch op.Type {
			case messages.OperationTypeInsert:
				m.status = fmt.Sprintf("Character inserted by User-%d", op.UserID)
			case messages.OperationTypeDelete:
				m.status = fmt.Sprintf("Character deleted by User-%d", op.UserID)
			}
		}
	case messages.MessageTypeSync:
		if msg.UserID != m.userID && msg.Document != nil {
			// Handle document sync
			m.doc = msg.Document
			m.status = fmt.Sprintf("Document synchronized with User-%d", msg.UserID)
		}
	}
}

func (m *model) View() string {
	// Lipgloss styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("8"))
	highlightStyle := lipgloss.NewStyle().Reverse(true)
	notesStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1).
		MarginTop(1).
		BorderForeground(lipgloss.Color("8"))

	// Build text area
	var textLines []string
	maxLineLen := 0
	for y, line := range m.doc.Lines {
		var lineStr string
		for x, char := range line.Characters {
			highlight := false
			if m.selectionActive {
				// Selection is from (selStartY, selStartX) to (cursorY, cursorX)
				sy, sx := m.selStartY, m.selStartX
				ey, ex := m.cursorY, m.cursorX
				// Normalize selection order
				if sy > ey || (sy == ey && sx > ex) {
					sy, sx, ey, ex = ey, ex, sy, sx
				}
				// Selection is inclusive of start, exclusive of end
				if (y+1 > sy && y+1 < ey) ||
					(y+1 == sy && y+1 == ey && x+1 >= sx && x+1 < ex) ||
					(y+1 == sy && y+1 != ey && x+1 >= sx) ||
					(y+1 == ey && y+1 != sy && x+1 < ex) {
					highlight = true
				}
			}
			if m.cursorY == y+1 && m.cursorX == x+1 {
				lineStr += "_"
			}
			if highlight {
				lineStr += highlightStyle.Render(string(char.Value))
			} else {
				lineStr += string(char.Value)
			}
		}
		// Show cursor at end of line
		if m.cursorY == y+1 && m.cursorX == len(line.Characters)+1 {
			lineStr += "_"
		}
		if len(lineStr) > maxLineLen {
			maxLineLen = len(lineStr)
		}
		textLines = append(textLines, lineStr)
	}
	// Pad lines to same length for border
	for i := range textLines {
		if len(textLines[i]) < maxLineLen {
			textLines[i] += repeatRune(" ", maxLineLen-len(textLines[i]))
		}
	}
	textArea := borderStyle.Render(lipgloss.JoinVertical(lipgloss.Left, textLines...))

	// Build notes/commands area with fixed width
	notes := []string{
		fmt.Sprintf("Status: %s", m.status),
		"Commands:",
		"  Arrows: Move   Shift+Arrows: Select   Esc: Clear Selection",
		"  Type: Insert   Backspace/Delete: Delete   Enter: Newline",
		"  Ctrl+S: Save   Ctrl+Q: Quit",
	}
	notesBlock := notesStyle.Render(lipgloss.JoinVertical(lipgloss.Left, notes...))

	return textArea + "\n" + notesBlock
}

func repeatRune(s string, count int) string {
	if count <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// deleteSelection deletes the currently selected text region
func (m *model) deleteSelection() {
	if !m.selectionActive {
		return
	}
	// Normalize selection order
	sy, sx := m.selStartY, m.selStartX
	ey, ex := m.cursorY, m.cursorX
	if sy > ey || (sy == ey && sx > ex) {
		sy, sx, ey, ex = ey, ex, sy, sx
	}
	// Delete from end to start to avoid messing up positions
	for y := ey; y >= sy; y-- {
		line := m.doc.Lines[y-1]
		startX := 1
		endX := len(line.Characters)
		if y == sy {
			startX = sx
		}
		if y == ey {
			endX = ex - 1
		}
		for x := endX; x >= startX; x-- {
			if x-1 < 0 || x-1 >= len(m.doc.Lines[y-1].Characters) {
				continue
			}
			pos := m.doc.Lines[y-1].Characters[x-1].Pos
			_ = m.doc.DeleteCharacter(pos)
			m.sendDeleteOperation(pos)
		}
	}
	// Move cursor to start of selection
	m.cursorX = sx
	m.cursorY = sy
}

func StartTUI(editorState *shared.EditorState, userID int, userColor string) error {
	// Create model as a pointer to preserve program reference
	m := initialModel(editorState, userID, userColor)
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Store the program reference for message handling
	m.program = p

	return p.Start()
}

// Testing helpers

// MockModel is a wrapper around the model struct for testing purposes
type MockModel struct {
	*model
}

// InitializeModelForTesting creates a model for testing purposes
func InitializeModelForTesting(editorState *shared.EditorState, userID int, userColor string) *MockModel {
	return &MockModel{
		model: initialModel(editorState, userID, userColor),
	}
}

// GetDocumentText returns the document text for testing
func (m *MockModel) GetDocumentText() string {
	return m.doc.ToText()
}

// GetCursorPosition returns the cursor position for testing
func (m *MockModel) GetCursorPosition() (int, int) {
	return m.cursorX, m.cursorY
}

// SetCursorPosition sets the cursor position for testing
func (m *MockModel) SetCursorPosition(x, y int) {
	m.cursorX = x
	m.cursorY = y
}

// SimulateKeyPress simulates pressing a key for testing
func (m *MockModel) SimulateKeyPress(key string) {
	// Create a tea.KeyMsg and send it to Update
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	if key == "enter" {
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	} else if key == "backspace" {
		msg = tea.KeyMsg{Type: tea.KeyBackspace}
	} else if key == "left" {
		msg = tea.KeyMsg{Type: tea.KeyLeft}
	} else if key == "right" {
		msg = tea.KeyMsg{Type: tea.KeyRight}
	} else if key == "up" {
		msg = tea.KeyMsg{Type: tea.KeyUp}
	} else if key == "down" {
		msg = tea.KeyMsg{Type: tea.KeyDown}
	}

	newModel, _ := m.model.Update(msg)
	// Update should return the same model pointer
	if newModel != m.model {
		// In testing, we'll discard the commands and just ensure the model is updated
		*m.model = *(newModel.(*model))
	}
}
