package main

import (
	"fmt"
	"io"
	"log"
)

func (s *session) handleTest(r io.Reader) error {

	log.Print("test email received! generating reply...")

	if _, err := io.Copy(io.Discard, r); err != nil {
		return fmt.Errorf("error discarding body: %v", err)
	}

	replyTo := normalizeAddress(s.from)
	if replyTo == "" {
		return nil
	}

	body := "This is a test response!"
	filepath := "./.gitignore"

	err := sendEmail([]string{replyTo}, body, filepath)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
