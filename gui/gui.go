package gui

import (
	"fmt"
	"net"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type EditorChange struct {
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
	w := a.NewWindow("Client")

	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder("Type here...")

	entry.OnChanged = func(text string) {
		line, col := getCursorPosition(entry)
		fmt.Printf("Character inserted at Line %d, Column %d\n", line, col)
	}

	entry.OnCursorChanged = func() {
		line, col := getCursorPosition(entry)
		fmt.Printf("Cursor moved to Line %d, Column %d\n", line, col)
	}

	content := container.NewVBox(entry)
	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()
}
