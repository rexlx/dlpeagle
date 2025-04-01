package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type Storage interface {
	SavePDF(file []byte, name string) error
	SaveImage(file []byte, name string) error
	SaveHTML(file []byte, name string) error
	// GetPDF(name string) ([]byte, error)
	// GetImage(name string) ([]byte, error)
	// GetHTML(name string) ([]byte, error)
	// DeletePDF(name string) error
	// DeleteImage(name string) error
	// DeleteHTML(name string) error
	ListPDFs() ([]string, error)
	ListImages() ([]string, error)
	ListHTMLs() ([]string, error)
}

type S3Storage struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
	UsePath   bool
	UseIAM    bool
}

type HttpStorage struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

type LocalStorage struct {
	Root    string
	UsePath bool
}

type MinioStorage struct{}

func (h *HttpStorage) saveFile(file []byte, name, fileType string) error {
	chunkSize := 1024 * 1024 // 1 MB
	url := h.Endpoint + "/upload"
	var lastChunk bool
	client := &http.Client{}
	for i := 0; i < len(file); i += chunkSize {
		end := i + chunkSize
		if end > len(file) {
			end = len(file)
			lastChunk = true
		}
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Authorization", "AWS "+h.AccessKey+":"+h.SecretKey)
		req.Header.Set("X-filename", name)
		if lastChunk {
			req.Header.Set("X-Last-Chunk", "true")
		}
		req.Body = io.NopCloser(bytes.NewReader(file[i:end]))
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to upload chunk: %s", resp.Status)
		}
	}
	return nil
}

func (h *HttpStorage) SavePDF(file []byte, name string) error {
	return h.saveFile(file, name, "pdf")
}

func (h *HttpStorage) SaveImage(file []byte, name string) error {
	return h.saveFile(file, name, "image")
}

func (h *HttpStorage) SaveHTML(file []byte, name string) error {
	return h.saveFile(file, name, "html")
}

func (h *HttpStorage) getFile(name, fileType string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *HttpStorage) deleteFile(name, fileType string) error {
	return fmt.Errorf("not implemented")
}

func (h *HttpStorage) listFiles(fileType string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *HttpStorage) ListPDFs() ([]string, error) {
	return h.listFiles("pdf")
}

func (h *HttpStorage) ListImages() ([]string, error) {
	return h.listFiles("image")
}

func (h *HttpStorage) ListHTMLs() ([]string, error) {
	return h.listFiles("html")
}

func (l *LocalStorage) saveFile(file []byte, name, fileType string) error {
	dir := filepath.Join(l.Root, fileType)
	if l.UsePath {
		dir = l.Root
	}
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, name)
	return ioutil.WriteFile(filePath, file, 0644)
}

func (l *LocalStorage) getFile(name, fileType string) ([]byte, error) {
	dir := filepath.Join(l.Root, fileType)
	if l.UsePath {
		dir = l.Root
	}
	filePath := filepath.Join(dir, name)
	return ioutil.ReadFile(filePath)
}

func (l *LocalStorage) deleteFile(name, fileType string) error {
	dir := filepath.Join(l.Root, fileType)
	if l.UsePath {
		dir = l.Root
	}
	filePath := filepath.Join(dir, name)
	return os.Remove(filePath)
}

func (l *LocalStorage) listFiles(fileType string) ([]string, error) {
	dir := filepath.Join(l.Root, fileType)
	if l.UsePath {
		dir = l.Root
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return fileNames, nil
}

func (l *LocalStorage) SavePDF(file []byte, name string) error {
	return l.saveFile(file, name, "pdf")
}

func (l *LocalStorage) SaveImage(file []byte, name string) error {
	return l.saveFile(file, name, "image")
}

func (l *LocalStorage) SaveHTML(file []byte, name string) error {
	return l.saveFile(file, name, "html")
}

func (l *LocalStorage) GetPDF(name string) ([]byte, error) {
	return l.getFile(name, "pdf")
}

func (l *LocalStorage) GetImage(name string) ([]byte, error) {
	return l.getFile(name, "image")
}

func (l *LocalStorage) GetHTML(name string) ([]byte, error) {
	return l.getFile(name, "html")
}

func (l *LocalStorage) DeletePDF(name string) error {
	return l.deleteFile(name, "pdf")
}

func (l *LocalStorage) DeleteImage(name string) error {
	return l.deleteFile(name, "image")
}

func (l *LocalStorage) DeleteHTML(name string) error {
	return l.deleteFile(name, "html")
}

func (l *LocalStorage) ListPDFs() ([]string, error) {
	return l.listFiles("pdf")
}

func (l *LocalStorage) ListImages() ([]string, error) {
	return l.listFiles("image")
}

func (l *LocalStorage) ListHTMLs() ([]string, error) {
	return l.listFiles("html")
}
