package gui

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"gollaborate/crdt"
	"gollaborate/cursor"
	"gollaborate/messages"
)

type EditorState struct {
	document  *crdt.Document
	cursorMgr *cursor.Manager
	nodeID    int
	clock     int
	conns     []net.Conn // support multiple peer connections
	connMutex sync.Mutex // protects conns
	entry     *widget.Entry
	lastText  string
	updating  bool
}

func NewEditorState(conns []net.Conn, nodeID int) *EditorState {
	doc := crdt.FromText("", nodeID)
	cursorMgr := cursor.NewManager(doc, nodeID, "User", "#FF0000")

	return &EditorState{
		document:  doc,
		cursorMgr: cursorMgr,
		nodeID:    nodeID,
		clock:     1,
		conns:     conns,
		lastText:  "",
		updating:  false,
	}
}

// AddConn allows adding a new peer connection at runtime.
func (es *EditorState) AddConn(conn net.Conn) {
	es.connMutex.Lock()
	es.conns = append(es.conns, conn)
	es.connMutex.Unlock()
	go handleNetworkMessages(conn, es)
}

// SetNodeID allows updating the nodeID and cursorMgr after creation (for server-assigned IDs)
func (es *EditorState) SetNodeID(nodeID int) {
	es.nodeID = nodeID
	es.cursorMgr = cursor.NewManager(es.document, nodeID, "User", "#FF0000")
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

	// Apply operations to CRDT and broadcast to all peers
	for _, op := range operations {
		err := es.applyOperation(op)
		if err != nil {
			fmt.Printf("Error applying operation: %v\n", err)
			continue
		}

		// Broadcast operation to all connected peers (thread-safe)
		es.connMutex.Lock()
		for _, c := range es.conns {
			messages.SendOperation(c, op)
		}
		es.connMutex.Unlock()
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

// (syncWithServer removed - not needed in decentralized mode)

// handleRemoteOperation processes operations from other clients
func (es *EditorState) handleRemoteOperation(op *messages.Operation) error {
	// Don't apply our own operations
	if op.UserID == es.nodeID {
		return nil
	}

	// Update our clock to ensure we're synchronized
	if op.Clock > es.clock {
		es.clock = op.Clock
	}

	// Apply the operation to our local document
	err := es.applyCRDTOperation(op)
	if err != nil {
		return fmt.Errorf("failed to apply remote operation: %w", err)
	}

	fmt.Printf("Applied remote %s operation from user %d\n", op.Type, op.UserID)
	return nil
}

// handleDocumentSync replaces local document with server's authoritative version
func (es *EditorState) handleDocumentSync(doc *crdt.Document) {
	es.updating = true
	defer func() { es.updating = false }()

	// Replace our document with the server's version
	es.document = doc
	es.cursorMgr.UpdateDocument(es.document)

	// Update GUI to reflect new document state
	newText := es.document.ToText()
	es.entry.SetText(newText)
	es.lastText = newText

	fmt.Println("Document synchronized with server")
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
	var conns []net.Conn
	if conn != nil {
		conns = append(conns, conn)
	}
	GuiWithPeers(conns, generatePeerID())
}

// GuiWithPeers launches the editor with a set of peer connections and a peer ID.
func GuiWithPeers(conns []net.Conn, peerID int, editorStateOpt ...*EditorState) {
	a := app.New()
	w := a.NewWindow("Collaborative Editor")

	var editorState *EditorState
	if len(editorStateOpt) > 0 && editorStateOpt[0] != nil {
		editorState = editorStateOpt[0]
	} else {
		editorState = NewEditorState(conns, peerID)
	}

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

		// Send cursor position over all peer connections (thread-safe)
		editorState.connMutex.Lock()
		for _, c := range editorState.conns {
			messages.SendCursor(c, position, editorState.nodeID, "User", "#FF0000")
		}
		editorState.connMutex.Unlock()

		// Track highlighted text
		highlighted := entry.SelectedText()
		if highlighted != "" {
			fmt.Printf("Highlighted text: '%s' (Line %d, Column %d)\n", highlighted, line, col)
			// TODO: Send selection update
		}
	}

	// Start network message handler for each peer connection
	for _, c := range editorState.conns {
		go handleNetworkMessages(c, editorState)
	}
	if len(editorState.conns) == 0 {
		// Initialize with empty document in offline mode
		editorState.updateGUIFromCRDT()
		fmt.Println("Started in offline mode")
	}

	content := container.NewVBox(entry)
	w.SetContent(content)
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}

// generatePeerID creates a unique peer ID for decentralized mode.
func generatePeerID() int {
	return int(time.Now().UnixNano() % 99999999)
}

// generateNodeID creates a unique node ID for this client
// generateNodeID is now unused in online mode; user ID is assigned by server.
// It is retained for offline mode only.
func generateNodeID(conn net.Conn) int {
	if conn == nil {
		return 1 // Default for offline mode
	}
	return 1 // Placeholder, not used in online mode
}

func handleNetworkMessages(conn net.Conn, editorState *EditorState) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Network handler crashed: %v\n", r)
		}
	}()

	for {
		msg, err := messages.ReceiveMessage(conn)
		if err != nil {
			fmt.Printf("Connection lost: %v\n", err)
			return
		}

		switch msg.Type {
		case messages.MessageTypeOperation:
			err := editorState.handleRemoteOperation(msg.Operation)
			if err != nil {
				fmt.Printf("Error handling remote operation: %v\n", err)
			}
		case messages.MessageTypeCursor:
			if msg.Cursor != nil && msg.Cursor.UserID != editorState.nodeID {
				handleRemoteCursor(msg.Cursor)
			}
		case messages.MessageTypeSelection:
			if msg.Selection != nil && msg.Selection.UserID != editorState.nodeID {
				handleRemoteSelection(msg.Selection)
			}
		case messages.MessageTypeAck:
			fmt.Printf("Peer acknowledged operation\n")
		case messages.MessageTypeError:
			fmt.Printf("Peer error: %s\n", msg.Error)
		default:
			fmt.Printf("Unknown message type: %s\n", msg.Type)
		}
	}
}

// handleRemoteCursor displays cursor position from other users
func handleRemoteCursor(cursor *messages.CursorPosition) {
	fmt.Printf("User %s (%s) cursor at position %v\n",
		cursor.UserName, cursor.Color, cursor.Position)
	// TODO: Display cursor in GUI
}

// handleRemoteSelection displays selection from other users
func handleRemoteSelection(selection *messages.Selection) {
	if selection.StartPosition == nil && selection.EndPosition == nil {
		fmt.Printf("User %s (%s) cleared selection\n",
			selection.UserName, selection.Color)
	} else {
		fmt.Printf("User %s (%s) selected from %v to %v\n",
			selection.UserName, selection.Color,
			selection.StartPosition, selection.EndPosition)
	}
	// TODO: Display selection in GUI
}
