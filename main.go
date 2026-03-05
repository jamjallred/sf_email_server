package main

import (
	"errors"
	"io"
	"log"
	"net/mail"
	"os"
	"strings"

	smtplib "github.com/emersion/go-smtp"
	"github.com/joho/godotenv"
)

// Configurable bits
const (
	serverDomain  = "mail.soonerfleet.com" // the hostname of this server
	listenAddress = ":25"                  // must be 25 for normal SMTP
	outboundHost  = "mail.soonerfleet.com"
	outboundAddr  = outboundHost + ":25"
)

// whitelist for endpoints
var allowedEndpoints = make(map[string]bool)
var allowedSenders = make(map[string]bool)

func load_config() {

	val := os.Getenv("ALLOWED_ENDPOINTS")
	for _, item := range strings.Split(val, ",") {
		cleanItem := strings.ToLower(strings.TrimSpace(item))
		if cleanItem != "" {
			allowedEndpoints[cleanItem] = true
		}
	}

	val = os.Getenv("ALLOWED_SENDERS")
	for _, item := range strings.Split(val, ",") {
		cleanItem := strings.ToLower(strings.TrimSpace(item))
		if cleanItem != "" {
			allowedSenders[cleanItem] = true
		}
	}
}

// backend implements the SMTP server backend.
type backend struct{}

func (b *backend) NewSession(c *smtplib.Conn) (smtplib.Session, error) {
	log.Printf("New connection from %s", c.Conn().RemoteAddr())
	return &session{}, nil
}

// session stores state for a single SMTP session.
type session struct {
	from string
	to   []string
}

func (s *session) Mail(from string, opts *smtplib.MailOptions) error {
	addr := normalizeAddress(from)
	log.Printf("MAIL FROM: %s", addr)

	if allowedSenders[addr] {
		s.from = addr
		return nil
	}

	log.Printf("Rejecting MAIL: %s is not a recognized sender", addr)
	return errors.New("550 5.7.1 sender not authorized")
}

func (s *session) Rcpt(to string, opts *smtplib.RcptOptions) error {
	addr := strings.ToLower(normalizeAddress(to))
	log.Printf("RCPT TO: %s", addr)

	// whitelisting endpoints
	if !allowedEndpoints[addr] {
		log.Printf("Rejecting RCPT for non-allowed recipient: %s", addr)
		return errors.New("550 5.7.1 unsupported endpoint")
	}

	s.to = append(s.to, addr)
	return nil
}

func (s *session) Data(r io.Reader) error {

	log.Print("processing incoming email!")

	if containsAddr(s.to, "generate@soonerfleet.com") {
		err := s.handleGenerate(r)
		if err != nil {
			log.Printf("Error in handleGenerate(): %s", err)
		}
		return nil

	}

	if containsAddr(s.to, "test@soonerfleet.com") {
		err := s.handleTest(r)
		if err != nil {
			log.Printf("Error in handleTest(): %s", err)
		}
		return nil

	}

	if containsAddr(s.to, "savetodb@soonerfleet.com") {
		err := s.handleSaveToDB(r)
		if err != nil {
			log.Printf("Error in handleSaveToDB(): %s", err)
		}
		return nil
	}

	return nil

}

func (s *session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *session) Logout() error {
	return nil
}

// normalizeAddress takes something like "<User <u@example.com>>" and returns "u@example.com".
func normalizeAddress(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "<>")
	// Try to parse with net/mail to be robust:
	if addr, err := mail.ParseAddress(raw); err == nil {
		return strings.ToLower(strings.TrimSpace(addr.Address))
	}
	return strings.ToLower(raw)
}

func main() {
	godotenv.Load()
	load_config()

	be := &backend{}
	s := smtplib.NewServer(be)

	s.Addr = listenAddress
	s.Domain = serverDomain
	s.AllowInsecureAuth = true // fine for internal / no-auth use
	// Optional timeouts, limits, etc.:
	// s.ReadTimeout = 10 * time.Second
	// s.WriteTimeout = 10 * time.Second
	// s.MaxMessageBytes = 10 << 20 // 10 MB

	log.Printf("Starting SMTP server on %s for domain %s", s.Addr, s.Domain)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func containsAddr(addrs []string, target string) bool {
	target = strings.ToLower(normalizeAddress(target))
	for _, a := range addrs {
		if strings.ToLower(normalizeAddress(a)) == target {
			return true
		}
	}
	return false
}
