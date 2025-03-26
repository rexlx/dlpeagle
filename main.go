package main

import (
	"fmt"
	"os"

	// "fyne.io/fyne/canvas"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	api := API{
		URL:      "http://localhost:8081",
		Username: "admin",
		Password: "password",
	}
	instance := NewInstance(api, "instance.log")
	instance.Logger.Println("Starting application...")
	a := app.New()
	w := a.NewWindow("Document Metadata Tool")

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentIcon(), func() {
			fmt.Println("Document action clicked")
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			fmt.Println("Settings clicked")
		}),
	)
	resource, err := fyne.LoadResourceFromPath("data/bg.jpg")
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	bgImage := canvas.NewImageFromResource(resource)
	bgImage.FillMode = canvas.ImageFillStretch
	content := widget.NewLabel("Drag and drop a document here.")
	// stackedContent := container.NewStack(bgImage, container.NewCenter(content))
	stackedContent := container.NewStack(
		bgImage,
		container.NewCenter(content), // Center the label
	)

	w.SetContent(container.NewBorder(toolbar, nil, nil, nil, stackedContent))

	w.Resize(fyne.NewSize(600, 400))

	w.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		if len(uris) == 0 {
			return
		}

		filePath := uris[0].Path()
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		if fileInfo.IsDir() {
			dialog.ShowInformation("Error", "Directories are not supported.", w)
			return
		}

		documentType := instance.inferDocumentType(filePath)
		metadata := generateMetadata(filePath, documentType)

		// Display results (replace with your metadata handling logic)
		resultText := fmt.Sprintf("File: %s\nType: %s\nMetadata: %v", filePath, documentType, metadata)
		content.SetText(resultText)
	})

	w.ShowAndRun()
}
