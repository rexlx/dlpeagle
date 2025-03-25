package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
)

func main() {
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

	content := widget.NewLabel("Drag and drop a document here.")

	w.SetContent(container.NewBorder(toolbar, nil, nil, nil, content))

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

		documentType := inferDocumentType(filePath)
		metadata := generateMetadata(filePath, documentType)

		// Display results (replace with your metadata handling logic)
		resultText := fmt.Sprintf("File: %s\nType: %s\nMetadata: %v", filePath, documentType, metadata)
		content.SetText(resultText)
	})

	w.ShowAndRun()
}

func inferDocumentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return "PDF"
	case ".txt":
		return "Text"
	case ".docx", ".doc":
		return "Word Document"
	case ".jpg", ".jpeg", ".png", ".gif":
		return "Image"
	default:
		return "Unknown"
	}
}

func generateMetadata(filePath string, documentType string) map[string]string {
	metadata := make(map[string]string)

	// Add basic metadata. Enhance this with actual metadata extraction.
	metadata["filename"] = filepath.Base(filePath)
	metadata["type"] = documentType

	fileInfo, err := os.Stat(filePath)
	if err == nil {
		metadata["size"] = fmt.Sprintf("%d bytes", fileInfo.Size())
	}

	//Add further meta data as needed.
	return metadata
}

type Tag struct {
	Username string `json:"username"`
	FilePath string `json:"file_path"`
	ID       string `json:"id"`
	ClientID string `json:"client_id"`
	Hash     string `json:"hash"`
	URL      string `json:"url"`
	Created  int    `json:"created"`
}

func (i *Instance) TagWordDocument(filePath string) {
	if !FileExists(filePath) {
		fmt.Println("File does not exist.")
		return
	}
	newLine := `<w:instrText xml:space="preserve"> INCLUDEPICTURE \d "%v" \* MERGEFORMATINET </w:instrText>`
	//add newline to word document
	err := addLineToWordDocument(filePath, newLine)
	if err != nil {
		fmt.Println("Error adding line to word document:", err)
		return
	}
	hash, err := CalculateSHA256(filePath)
	if err != nil {
		fmt.Println("Error calculating hash:", err)
		return
	}
	var t Tag
	t.FilePath = filePath
	t.ID = uuid.New().String()
	t.Created = int(time.Now().Unix())
	t.Hash = hash
	out, err := json.Marshal(t)
	if err != nil {
		fmt.Println("Error marshalling tag:", err)
		return
	}
	request, err := http.NewRequest("POST", i.API.URL, bytes.NewBuffer(out))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	request.SetBasicAuth(i.API.Username, i.API.Password)
	res, err := i.Gateway.Do(request)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}
	fmt.Println(string(body))
	// send this to an api eventually
}
