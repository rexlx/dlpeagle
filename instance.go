package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Instance struct {
	API     API          `json:"api"`
	Logger  *log.Logger  `json:"-"`
	Gateway *http.Client `json:"-"`
}

type API struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewInstance(api API, logname string) *Instance {
	f, err := os.OpenFile(logname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	logger := log.New(f, "instance: ", log.LstdFlags)
	logger.Println("Creating new instance...")
	return &Instance{
		Logger:  logger,
		API:     api,
		Gateway: http.DefaultClient,
	}
}

func (i *Instance) SendTag(tag Tag) error {
	return nil
}

func (i *Instance) inferDocumentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return "PDF"
	case ".txt":
		return "Text"
	case ".docx", ".doc":
		i.TagWordDocument(filePath)
		return "Word Document"
	case ".jpg", ".jpeg", ".png", ".gif":
		return "Image"
	default:
		return "Unknown"
	}
}

func (i *Instance) TagWordDocument(filePath string) {
	if !FileExists(filePath) {
		fmt.Println("File does not exist.")
		return
	}
	// TODO: should we hash file before and after?
	id := uuid.New().String()
	url := fmt.Sprintf("%v/%v", i.API.URL, id)
	// add the hidden image to the word document
	t, err := addRemoteImageTrackerToWordDocument(filePath, url, id)
	if err != nil {
		i.Logger.Println("Error adding line to word document:", err)
		return
	}
	// err = i.SendTag(t)

	hash, err := CalculateSHA256(filePath)
	if err != nil {
		i.Logger.Println("Error calculating hash:", err)
		return
	}
	t.Hash = hash
	uname, err := GetUsername()
	if err != nil {
		i.Logger.Println("Error getting username:", err)
	}
	t.Username = uname
	t.FilePath = filePath
	out, err := json.Marshal(t)
	if err != nil {
		i.Logger.Println("Error marshalling tag:", err)
		return
	}
	request, err := http.NewRequest("POST", fmt.Sprintf("%v/tag", i.API.URL), bytes.NewBuffer(out))
	if err != nil {
		i.Logger.Println("Error creating request:", err)
		return
	}
	request.SetBasicAuth(i.API.Username, i.API.Password)
	res, err := i.Gateway.Do(request)
	if err != nil {
		i.Logger.Println("Error sending request:", err)
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		i.Logger.Println("Error reading response body:", err)
		return
	}
	i.Logger.Println(string(body))
	// send this to an api eventually
}
