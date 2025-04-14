package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

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
		URL:      "http://fairlady:8081",
		Username: "admin",
		Password: "password",
	}
	f, err := os.OpenFile("instance.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	logger := log.New(f, "instance: ", log.LstdFlags)
	storage := HttpStorage{
		Endpoint: api.URL,
	}
	messageLabel := widget.NewLabel("")
	instance := NewInstance(api, logger, "localhost:4242", messageLabel)
	instance.Storage = &storage
	instance.Logger.Println("Starting application...")
	a := app.NewWithID("com.example.dlpeagle")
	w := a.NewWindow("DLPeagle")
	instance.Window = w

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentIcon(), func() {
			fmt.Println("testing server connection")
			if !instance.IsConnected() {
				dialog.ShowInformation("Connection", "Not connected to the server.", w)
				return
			}
			dialog.ShowInformation("Connection", "Connected to the server.", w)
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

	// Create a warning rectangle (initially hidden)
	warningRect := canvas.NewRectangle(color.RGBA{255, 0, 0, 128}) // Semi-transparent red
	warningRect.Hide()

	stackedContent := container.NewStack(
		bgImage,
		container.NewCenter(content),
		warningRect, // Add the warning rectangle
	)

	w.SetContent(container.NewBorder(toolbar, nil, nil, nil, stackedContent))

	// w.Resize(fyne.NewSize(600, 400))
	w.Resize(fyne.NewSize(800, 600))

	w.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		if len(uris) == 0 {
			return
		}
		if !instance.IsConnected() {
			// Display warning
			warningRect.Show()
			warningRect.Resize(fyne.NewSize(w.Canvas().Size().Width, 20)) // Adjust size as needed
			warningRect.Move(fyne.NewPos(0, 0))                           // Position at the top

			// Hide warning after a delay (e.g., 5 seconds)
			time.AfterFunc(5*time.Second, func() {
				warningRect.Hide()
			})
		} else {
			// If connected, ensure the warning is hidden
			warningRect.Hide()
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

	// Initial API connection check
	if !instance.IsConnected() {
		warningRect.Show()
		warningRect.Resize(fyne.NewSize(w.Canvas().Size().Width, 20))
		warningRect.Move(fyne.NewPos(0, 0))
		time.AfterFunc(5*time.Second, func() {
			warningRect.Hide()
		})
	}

	w.ShowAndRun()
}
