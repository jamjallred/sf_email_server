package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"

	excelutils "github.com/jamjallred/sf_server_utils"
	"github.com/joho/godotenv"
)

func (s *session) handleGenerate(r io.Reader) error {
	godotenv.Load()
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
	//var data []byte
	for {
		part, err := mr.NextPart()
		fmt.Println("email part:")
		fmt.Println(part) // TESTING LINE ````````````````````````````````````````
		if err == io.EOF {
			fmt.Println("no more parts")
			break
		}

		cd := part.Header.Get("Content-Disposition")
		if !strings.Contains(cd, ".xlsx") {
			continue
		}

		cte := strings.ToLower(part.Header.Get("Content-Transfer-Encoding"))

		if cte != "base64" {
			return fmt.Errorf("unknown attachment encoding: %s", cte)
		}
		src := base64.NewDecoder(base64.StdEncoding, part)

		// we have an excel file attachment and it's the current part
		// save this part to disk
		xlsxPath = filepath.Join(os.TempDir(), "temp_excel_file.xlsx")
		f, err := os.Create(xlsxPath)
		if err != nil {
			fmt.Println("error creating xlsxPath")
		}
		bytes, err := io.Copy(f, src)
		//data, err = io.ReadAll(part)
		if err != nil {
			fmt.Println("error saving file to disk")
		}
		fmt.Printf("Wrote %v bytes to file\n", bytes)
		f.Close()
	}

	log.Println("File successfully saved at: ", xlsxPath)
	savePath := "./assets/" + os.Getenv("FILENAME_PREFIX") + time.Now().Format("2006-01-02") + ".xlsx"
	fmt.Printf("This is the savepath:\n %s\n", savePath)

	// Generate the excel file to attach to the email
	if err := excelutils.Generate(xlsxPath, savePath); err != nil {
		log.Printf("unable to generate xlsx file")
		return err
	}

	// Send the email

	recipients := []string{
		"stancoppinger@tulsacoxmail.com",
		"mike.allred@tulsacoxmail.com",
		"gregg.wessels@tulsacoxmail.com",
		"jaxcoppinger@tulsacoxmail.com",
		"jj.soonerfleet@gmail.com",
	}

	if err := saveToDB(xlsxPath); err != nil {
		fmt.Printf("error saving to database %v\n", err)
		return err
	}

	body := "Spreadsheet generated successfully!"
	if err := sendEmail(recipients, body, savePath); err != nil {
		fmt.Println("unable to send email: ", err)
		os.Remove(savePath)
		os.Remove(xlsxPath)
		return err
	}

	os.Remove(savePath)
	os.Remove(xlsxPath)

	return nil
}

func saveToDB(xlsxPath string) error {

	godotenv.Load()

	if _, err := os.Stat(xlsxPath); err != nil {
		fmt.Println("something has gone terribly wrong")
		return err
	}

	if len(xlsxPath) <= 5 || xlsxPath[len(xlsxPath)-5:] != ".xlsx" {
		return errors.New("file must be an .xlsx")
	}

	f, err := os.Open(xlsxPath)
	if err != nil {
		return err
	}
	defer f.Close()

	url := os.Getenv("WEB_ENDPOINT_SAVE")
	resp, err := http.Post(url, "application/octet-stream", f)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	log.Println("Information successfully saved to database!")

	return nil
}
