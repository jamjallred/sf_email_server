package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/mailgun/mailgun-go/v4"
)

func sendEmail(recipients []string, body, filepath string) error {
	godotenv.Load()
	domain := "mail.soonerfleet.com"
	apiKey := os.Getenv("MAILGUN_API_KEY")
	if apiKey == "" {
		log.Fatal("MAILGUN_API_KEY not set")
	}

	mg := mailgun.NewMailgun(domain, apiKey)

	sender := "no-reply@" + domain
	subject := os.Getenv("FILENAME_PREFIX") + time.Now().Format("2006-01-02")

	message := mailgun.NewMessage(sender, subject, body, recipients...)
	message.AddAttachment(filepath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, id, err := mg.Send(ctx, message)
	if err != nil {
		return err
	}

	log.Println("email sent, ID: ", id)
	return nil
}
