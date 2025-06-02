package gui

import (
	"fmt"
	"net"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	
	"gollaborate/crdt"
	"gollaborate/cursor"
	"gollaborate/messages"
)

type EditorState struct {
	document    *crdt.Document
	cursorMgr   *cursor.Manager
	nodeID      int
	clock       int
	conn        net.Conn
	entry       *widget.Entry
	lastText    string
	updating    bool
}

func NewEditorState(conn net.Conn, nodeID int) *EditorState {
	doc := crdt.FromText("", nodeID)
	cursorMgr := cursor.NewManager(doc, nodeID, "User", "#FF0000")
	
	return &EditorState{
		document:  doc,
		cursorMgr: cursorMgr,
		nodeID:    nodeID,
		clock:     1,
		conn:      conn,
		lastText:  "",
		updating:  false,
	}
}

func (es *EditorState) nextClock() int {
	es.clock++
	return es.clock
}

func (es *EditorState) updateGUIFromCRDT() {
	if es.updating {
		return
	}
	
	es.updating = true
	defer func() { es.updating = false }()
	
	newText := es.document.ToText()
	es.entry.SetText(newText)
	es.lastText = newText
}

func (es *EditorState) processTextChange(newText string) {
	if es.updating {
		return
	}
	
	// Find differences between old and new text
	operations := es.detectChanges(es.lastText, newText)
	
	// Apply operations to CRDT
	for _, op := range operations {
		err := es.applyOperation(op)
		if err != nil {
			fmt.Printf("Error applying operation: %v\n", err)
			continue
		}
		
		// Send operation over network
		if es.conn != nil {
			messages.SendOperation(es.conn, op)
		}
	}
	
	es.lastText = newText
	es.cursorMgr.UpdateDocument(es.document)
}

func (es *EditorState) detectChanges(oldText, newText string) []*messages.Operation {
	var operations []*messages.Operation
	
	// Simple diff algorithm - this could be improved
	oldRunes := []rune(oldText)
	newRunes := []rune(newText)
	
	i, j := 0, 0
	line, col := 1, 1
	
	for i < len(oldRunes) || j < len(newRunes) {
		if i < len(oldRunes) && j < len(newRunes) && oldRunes[i] == newRunes[j] {
			// Characters match
			if oldRunes[i] == '\n' {
				line++
				col = 1
			} else {
				col++
			}
			i++
			j++
		} else if j < len(newRunes) && (i >= len(oldRunes) || oldRunes[i] != newRunes[j]) {
			// Character inserted
			position, err := es.document.GeneratePositionAt(line, col, es.nodeID)
			if err != nil {
				fmt.Printf("Error generating position: %v\n", err)
				j++
				continue
			}
			
			op := messages.NewInsertOperation(position, newRunes[j], es.nodeID, es.nextClock())
			operations = append(operations, op)
			
			if newRunes[j] == '\n' {
				line++
				col = 1
			} else {
				col++
			}
			j++
		} else if i < len(oldRunes) {
			// Character deleted
			position, err := es.document.FindPositionAt(line, col)
			if err != nil {
				fmt.Printf("Error finding position for deletion: %v\n", err)
				i++
				continue
			}
			
			op := messages.NewDeleteOperation(position, es.nodeID, es.nextClock())
			operations = append(operations, op)
			
			i++
		}
	}
	
	return operations
}

func (es *EditorState) applyOperation(op *messages.Operation) error {
	switch op.Type {
	case messages.OperationTypeInsert:
		return es.document.InsertCharacter(op.Character, op.Position, op.Clock)
	case messages.OperationTypeDelete:
		return es.document.DeleteCharacter(op.Position)
	default:
		return fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

func (es *EditorState) applyCRDTOperation(op *messages.Operation) error {
	err := es.applyOperation(op)
	if err != nil {
		return err
	}
	
	es.cursorMgr.UpdateDocument(es.document)
	es.updateGUIFromCRDT()
	return nil
}

func getCursorPosition(entry *widget.Entry) (int, int) {
	cursorPos := entry.CursorColumn + entry.CursorRow*entry.CursorColumn
	if cursorPos > len(entry.Text) {
		cursorPos = len(entry.Text)
	}

	// Split the text into lines
	lines := strings.Split(entry.Text, "\n")

	// Calculate the line and column based on the cursor position
	line := entry.CursorRow + 1
	column := entry.CursorColumn + 1

	// Ensure the column number is within the bounds of the current line
	if line <= len(lines) {
		if column > len(lines[line-1]) {
			column = len(lines[line-1]) + 1
		}
	}

	return line, column
}

func Gui(conn net.Conn) {
	a := app.New()
	w := a.NewWindow("Collaborative Editor")

	// Initialize editor state
	editorState := NewEditorState(conn, 1) // TODO: Get actual node ID

	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder("Start typing...")
	editorState.entry = entry

	// Handle text changes
	entry.OnChanged = func(text string) {
		editorState.processTextChange(text)
	}

	// Handle cursor movements
	entry.OnCursorChanged = func() {
		line, col := getCursorPosition(entry)
		fmt.Printf("Cursor moved to Line %d, Column %d\n", line, col)

		// Convert to CRDT position and send cursor update
		position, err := editorState.cursorMgr.GetCRDTPositionFromTextCoords(line, col)
		if err != nil {
			fmt.Printf("Error getting CRDT position: %v\n", err)
			return
		}

		// Send cursor position over network
		if conn != nil {
			messages.SendCursor(conn, position, editorState.nodeID, "User", "#FF0000")
		}

		// Track highlighted text
		highlighted := entry.SelectedText()
		if highlighted != "" {
			fmt.Printf("Highlighted text: '%s' (Line %d, Column %d)\n", highlighted, line, col)
			// TODO: Send selection update
		}
	}

	// Start network message handler (if connection exists)
	if conn != nil {
		go handleNetworkMessages(conn, editorState)
	}

	// Initialize with empty document
	editorState.updateGUIFromCRDT()

	content := container.NewVBox(entry)
	w.SetContent(content)
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}

func handleNetworkMessages(conn net.Conn, editorState *EditorState) {
	for {
		msg, err := messages.ReceiveMessage(conn)
		if err != nil {
			fmt.Printf("Error receiving message: %v\n", err)
			return
		}

		switch msg.Type {
		case messages.MessageTypeOperation:
			if msg.Operation.UserID != editorState.nodeID {
				err := editorState.applyCRDTOperation(msg.Operation)
				if err != nil {
					fmt.Printf("Error applying remote operation: %v\n", err)
				}
			}
		case messages.MessageTypeInit:
			// Replace local document with server's document
			editorState.document = msg.Document
			editorState.cursorMgr.UpdateDocument(editorState.document)
			editorState.updateGUIFromCRDT()
		case messages.MessageTypeSync:
			// Handle sync messages
			fmt.Printf("Received sync from user %d\n", msg.UserID)
		case messages.MessageTypeCursor:
			// Handle cursor updates from other users
			if msg.Cursor.UserID != editorState.nodeID {
				fmt.Printf("User %s cursor at position %v\n", msg.Cursor.UserName, msg.Cursor.Position)
			}
		case messages.MessageTypeSelection:
			// Handle selection updates from other users
			if msg.Selection.UserID != editorState.nodeID {
				fmt.Printf("User %s selection: %v to %v\n", msg.Selection.UserName, msg.Selection.StartPosition, msg.Selection.EndPosition)
			}
		case messages.MessageTypeError:
			fmt.Printf("Server error: %s\n", msg.Error)
		default:
			fmt.Printf("Unknown message type: %s\n", msg.Type)
		}
	}
}