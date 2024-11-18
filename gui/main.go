package main

import (
	"fmt"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Create a new Fyne application
	myApp := app.New()
	myWindow := myApp.NewWindow("Fyne Text Input Example")

	// Create a multi-line text entry widget
	textInput := widget.NewMultiLineEntry()

	// Create a button to print the text input
	submitButton := widget.NewButton("Submit", func() {
		fmt.Println("Entered text:", textInput.Text)
	})

	// Create a container to hold the text input and button
	content := container.NewVBox(textInput, submitButton)

	// Set the content of the window
	myWindow.SetContent(content)

	// Show and run the application
	myWindow.ShowAndRun()
}
