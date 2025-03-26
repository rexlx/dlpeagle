package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/quic-go/quic-go"
)

func FileExists(fileHandle string) bool {
	info, err := os.Stat(fileHandle)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func addLineToWordDocument(filePath, newLine string) error {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer r.Close()

	var documentXML string
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			data, err := io.ReadAll(rc)
			if err != nil {
				return err
			}
			documentXML = string(data)
			rc.Close()
			break
		}
	}

	if documentXML == "" {
		return fmt.Errorf("document.xml not found")
	}

	bodyEnd := "</w:body>"
	insertPos := strings.LastIndex(documentXML, bodyEnd)

	if insertPos == -1 {
		return fmt.Errorf("</w:body> tag not found")
	}

	updatedXML := documentXML[:insertPos] + newLine + documentXML[insertPos:]

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, f := range r.File {
		fw, err := w.Create(f.Name)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		_, err = io.Copy(fw, rc)
		if err != nil {
			return err
		}
		rc.Close()
	}

	fw, err := w.Create("word/document.xml")
	if err != nil {
		return err
	}

	_, err = fw.Write([]byte(updatedXML))
	if err != nil {
		return err
	}

	w.Close()

	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, buf)
	return err
}

func addRemoteImageTrackerToWordDocument(filePath, trackerURL string, id string) (Tag, error) {
	// Open the .docx file (ZIP archive)
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return Tag{}, err
	}
	defer r.Close()

	var documentXML string
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return Tag{}, err
			}
			data, err := io.ReadAll(rc)
			if err != nil {
				return Tag{}, err
			}
			documentXML = string(data)
			rc.Close()
			break
		}
	}

	if documentXML == "" {
		return Tag{}, fmt.Errorf("document.xml not found")
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromString(documentXML); err != nil {
		return Tag{}, err
	}
	elements := doc.FindElements("//w:instrText")
	for _, e := range elements {
		contains := strings.Contains(e.Text(), "INCLUDEPICTURE")
		fmt.Println(contains)
		if contains {
			id, ok := extractUUIDFromText(e.Text())
			fmt.Println(id, ok)
			if ok {
				return Tag{ID: id, FilePath: filePath}, nil
			}
		}
	}
	var t Tag
	t.FilePath = filePath
	t.ID = id
	t.Created = int(time.Now().Unix())
	// Construct the fiTd code with proper structure
	fieldCode := fmt.Sprintf(`
    <w:p>
        <w:r>
            <w:fldChar w:fldCharType="begin"/>
        </w:r>
        <w:r>
            <w:instrText xml:space="preserve"> INCLUDEPICTURE \d "%s" \* MERGEFORMATINET </w:instrText>
        </w:r>
        <w:r>
            <w:fldChar w:fldCharType="separate"/>
        </w:r>
        <w:r>
            <w:t> </w:t> <!-- Placeholder for the image; Word will fetch it -->
        </w:r>
        <w:r>
            <w:fldChar w:fldCharType="end"/>
        </w:r>
    </w:p>`, trackerURL)
	fmt.Println(fieldCode)

	// Insert the field code before </w:body>
	bodyEnd := "</w:body>"
	insertPos := strings.LastIndex(documentXML, bodyEnd)
	if insertPos == -1 {
		return Tag{}, fmt.Errorf("</w:body> tag not found")
	}
	updatedXML := documentXML[:insertPos] + fieldCode + documentXML[insertPos:]

	// Create new ZIP
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Copy all original files except document.xml
	for _, f := range r.File {
		fw, err := w.Create(f.Name)
		if err != nil {
			return Tag{}, err
		}
		rc, err := f.Open()
		if err != nil {
			return Tag{}, err
		}
		if f.Name == "word/document.xml" {
			_, err = fw.Write([]byte(updatedXML))
		} else {
			_, err = io.Copy(fw, rc)
		}
		if err != nil {
			return Tag{}, err
		}
		rc.Close()
	}

	// Finalize ZIP and overwrite original file
	err = w.Close()
	if err != nil {
		return Tag{}, err
	}
	outFile, err := os.Create(filePath)
	if err != nil {
		return Tag{}, err
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, buf)
	if err != nil {
		return Tag{}, err
	}
	return t, nil
}

func GetUsername() (string, error) {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERNAME"), nil // Windows
	}

	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.Username, nil // Unix-like systems
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

func extractUUIDFromText(text string) (string, bool) {
	// Regular expression for UUID (version 4)
	uuidRegex := regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}`)

	// Find the first UUID in the text
	uuidMatch := uuidRegex.FindString(text)

	if uuidMatch != "" {
		return uuidMatch, true
	}

	return "", false // UUID not found
}

func dialQUIC(url string, sm *SecretManager) (quic.Connection, quic.Stream, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
	defer cancel()

	conn, err := quic.DialAddr(ctx, url, sm.TC, sm.QC)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	return conn, stream, nil
}
