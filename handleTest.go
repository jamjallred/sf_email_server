package main

import (
	"fmt"
	"io"
	"log"
	"net/smtp"
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

	return sendPlainReply(
		"test@mail.soonerfleet.com",
		replyTo,
		"Test successful <subject>",
		"Test successful <body>",
	)
}

func sendPlainReply(from, to, subject, body string) error {
	log.Print("further generating plain reply...")
	msg := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=utf-8\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	log.Print("sending reply!")

	return smtp.SendMail(outboundAddr, nil, from, []string{to}, msg)
}
