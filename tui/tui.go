package core

import (
	"fmt"
	"sync"

	"gollaborate/crdt"
	"gollaborate/messages"
	"gollaborate/shared"

	tea "github.com/charmbracelet/bubbletea"
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
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left":
			// Handle cursor movement
			if m.cursorX > 1 {
				m.cursorX--
			}
		case "right":
			lineLen := 0
			if m.cursorY-1 < len(m.doc.Lines) {
				lineLen = len(m.doc.Lines[m.cursorY-1].Characters)
			}
			if m.cursorX <= lineLen {
				m.cursorX++
			}
		case "up":
			if m.cursorY > 1 {
				m.cursorY--
				lineLen := len(m.doc.Lines[m.cursorY-1].Characters)
				if m.cursorX > lineLen+1 {
					m.cursorX = lineLen + 1
				}
			}
		case "down":
			if m.cursorY < len(m.doc.Lines) {
				m.cursorY++
				lineLen := len(m.doc.Lines[m.cursorY-1].Characters)
				if m.cursorX > lineLen+1 {
					m.cursorX = lineLen + 1
				}
			}
		case "s":
			m.status = "Saved"
		case "backspace":
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
	s := "Gollaborate TUI (CRDT is ground truth)\n\n"
	for y, line := range m.doc.Lines {
		for x, char := range line.Characters {
			if m.cursorY == y+1 && m.cursorX == x+1 {
				s += "_"
			}
			s += string(char.Value)
		}
		// Show cursor at end of line
		if m.cursorY == y+1 && m.cursorX == len(line.Characters)+1 {
			s += "_"
		}
		s += "\n"
	}
	s += "\n\nStatus: " + m.status
	s += "\nArrows: Move | Type: Insert | Backspace: Delete | Enter: Newline | s: Save | q: Quit"
	return s
}

func StartTUI(editorState *shared.EditorState, userID int, userColor string) error {
	// Create model as a pointer to preserve program reference
	m := initialModel(editorState, userID, userColor)
	p := tea.NewProgram(m)

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
