package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/mailgun/mailgun-go/v4"
)

func (s *session) handleForward(r io.Reader, recipient string) error {

	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}

	bufWithCloser := io.NopCloser(buf)

	godotenv.Load()
	domain := os.Getenv("DOMAIN")
	apiKey := os.Getenv("MAILGUN_API_KEY")
	mg := mailgun.NewMailgun(domain, apiKey)

	recipients := []string{recipient}
	message := mailgun.NewMIMEMessage(bufWithCloser, recipients...)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, id, err := mg.Send(ctx, message)
	if err != nil {
		log.Printf("Mailgun forward error: %v", err)
		return err
	}

	log.Printf("Forward via Mailgun! ID: %s", id)
	return nil
}
