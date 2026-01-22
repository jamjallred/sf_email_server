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
	"time"

	excelutils "github.com/jamjallred/sf_server_utils"
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
		fmt.Println("email part:")
		fmt.Println(part) // TESTING LINE ````````````````````````````````````````
		if err == io.EOF {
			fmt.Println("no more parts")
			break
		}
		part.Header.Get("Content-Disposition")

		if !strings.Contains(part.Header.Get("Content-Disposition"), ".xlsx") {
			continue
		}

		// we have an excel file attachment and it's the current part
		// save this part to disk
		xlsxPath = filepath.Join(os.TempDir(), "temp_excel_file.xlsx")
		f, err := os.Create(xlsxPath)
		if err != nil {
			fmt.Println("error creating xlsxPath")
		}
		_, err = io.Copy(f, part)
		if err != nil {
			fmt.Println("error saving file to disk")
		}
	}

	log.Println("File successfully saved at: ", xlsxPath)
	savePath := "./assets/" + os.Getenv("FILENAME_PREFIX") + time.Now().Format("2006-01-02") + ".xlsx"

	// Generate the excel file to attach to the email
	if err := excelutils.Generate(xlsxPath, savePath); err != nil {
		log.Printf("unable to generate xlsx file")
		return err
	}

	// Send the email
	// recipients := []string{
	// 	"stancoppinger@tulsacoxmail.com",
	// 	"mike.allred@tulsacoxmail.com",
	// 	"gregg.wessels@tulsacoxmail.com",
	// 	"jaxcoppinger@tulsacoxmail.com",
	// 	"jj.soonerfleet@gmail.com",
	// }

	// body := "THIS IS A TEST EMAIL, DISREGARD" //Spreadsheet generated successfully!"
	// if err := sendEmail(recipients, body, savePath); err != nil {
	// 	fmt.Println("unable to send email: ", err)
	// 	os.Remove(savePath)
	os.Remove(xlsxPath)
	// 	return err
	// }

	// os.Remove(savePath)
	// os.Remove(xlsxPath)

	return nil
}
