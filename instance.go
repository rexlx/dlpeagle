package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/widget"
	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
)

type Instance struct {
	Storage       Storage         `json:"-"`
	Memory        *sync.RWMutex   `json:"-"`
	Notifications []Notification  `json:"notifications"`
	SM            *SecretManager  `json:"-"`
	Notifier      SoundBlock      `json:"notifier"`
	API           API             `json:"api"`
	Logger        *log.Logger     `json:"-"`
	Gateway       *http.Client    `json:"-"`
	QUICConn      quic.Connection `json:"-"`            // QUIC Connection.
	QUICStream    quic.Stream     `json:"-"`            // QUIC Stream.
	QUICAddress   string          `json:"quic_address"` // Address of the QUIC server.
	MessageLabel  *widget.Label   `json:"-"`            // Label to display messages.
}

type SecretManager struct {
	QC          *quic.Config
	TC          *tls.Config
	Destination net.Addr
}

type API struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewInstance(api API, logname *log.Logger, quicAddress string, messageLabel *widget.Label) *Instance {
	sb := SoundBlockIn880Hz(time.Second)
	return &Instance{
		Memory:        &sync.RWMutex{},
		Notifications: make([]Notification, 0),
		SM:            &SecretManager{},
		Notifier:      *sb,
		API:           api,
		Logger:        logname,
		Gateway:       &http.Client{},
		QUICAddress:   quicAddress,
		MessageLabel:  messageLabel,
	}
}

func (i *Instance) SendTag(tag Tag) error {
	return nil
}

func (i *Instance) inferDocumentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		i.HandlePDF(filePath)
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
	status := res.StatusCode
	if status != http.StatusOK || status != http.StatusCreated {
		i.Logger.Println("Error sending tag:", status)
		return
	}
	i.Logger.Println("Tag sent successfully.")

}

func (i *Instance) SendAndReceiveOverQUIC(ctx context.Context, url string, sm *SecretManager, qr *Notification) {
	out, err := json.Marshal(qr)
	if err != nil {
		i.Logger.Println("Error marshalling notification:", err)
		return
	}

	conn, stream, err := dialQUIC(url, sm)
	if err != nil {
		i.Logger.Printf("Failed to dial QUIC: %v", err)
		return
	}
	i.QUICConn = conn
	i.QUICStream = stream
	defer func() {
		if err := i.QUICConn.CloseWithError(0, "terminating"); err != nil {
			i.Logger.Printf("Error closing QUIC connection: %v", err)
		}
		// Stream will be closed automatically by connection closure
		// If you need explicit stream closure first, do it before CloseWithError
	}()
	_, err = stream.Write(out) // this request registers the client
	if err != nil {
		i.Logger.Println("Error writing to QUIC stream:", err)
		return
	}

	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := stream.Read(buf)
			if err != nil {
				if err == io.EOF {
					i.Logger.Println("QUIC stream closed by remote")
					return
				}
				i.Logger.Printf("Error reading from QUIC stream: %v", err)
				return
			}

			var not Notification
			if err := json.Unmarshal(buf[:n], &not); err != nil {
				i.Logger.Println("Error unmarshalling notification:", err)
				return
			}

			i.Memory.Lock()
			i.Notifications = append(i.Notifications, not)
			i.Memory.Unlock()
		}
	}
}

func (i *Instance) HandlePDF(filePath string) {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		i.Logger.Println("Error reading PDF file:", err)
		return
	}
	fileName := filepath.Base(filePath)
	err = i.Storage.SavePDF(fileData, fileName)
	if err != nil {
		i.Logger.Println("Error saving PDF file:", err)
		return
	}
	i.Logger.Println("PDF file saved successfully.")
}
