package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
)

func (s *session) handleGenerate(r io.Reader) error {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return fmt.Errorf("Failed to parse email: %v", err)
	}

	replyTo := normalizeAddress(s.from)
	if replyTo == "" {
		replyTo = normalizeAddress(msg.Header.Get("From"))
	}

	ct := msg.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return fmt.Errorf("error parsing content-type: %v", err)
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		log.Printf("no multipart body => no attachment")
		return nil
	}

	mr := multipart.NewReader(msg.Body, params["boundary"])
	var xlsxPath string
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read part: %v", err)
		}

		filename := part.FileName()
		if filename == "" || !strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
			continue
		}

		xlsxPath = filepath.Join(os.TempDir(), filename)
		f, err := os.Create(xlsxPath)
		defer os.Remove(xlsxPath)
		if err != nil {
			return fmt.Errorf("cannot create file: %v", err)
		}
		if _, err := io.Copy(f, part); err != nil {
			f.Close()
			return fmt.Errorf("cannot write attachment: %v", err)
		}
		f.Close()
		break
	}

	if xlsxPath == "" {
		log.Printf("no .xlsx attachment found")
		return nil
	}

	// call utils here

	return nil
}
