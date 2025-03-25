package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
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
