package main

import (
	"errors"
	"io"
	"log"
	"net/mail"
	"strings"

	smtplib "github.com/emersion/go-smtp"
)

// Configurable bits
const (
	serverDomain  = "mail.soonerfleet.com" // the hostname of this server
	listenAddress = ":25"                  // must be 25 for normal SMTP
	outboundHost  = "mail.soonerfleet.com"
	outboundAddr  = outboundHost + ":25"
)

// whitelist for endpoints
var allowedEndpoints = map[string]bool{
	"generate@soonerfleet.com": true,
	"reserve@soonerfleet.com":  true,
	"test@soonerfleet.com":     true,
}

// whitelist for senders
var allowedSenders = map[string]bool{
	"stancoppinger@tulsacoxmail.com": true,
	"mike.allred@tulsacoxmail.com":   true,
	"gregg.wessels@tulsacoxmail.com": true,
	"jaxcoppinger@tulsacoxmail.com":  true,
	"jj.soonerfleet@gmail.com":       true,
	"no-reply@mail.soonerfleet.com":  true,
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

	// whitelisting senders
	if strings.HasSuffix(strings.ToLower(addr), "@mail.soonerfleet.com") {
		s.from = addr
		return nil
	}

	if len(allowedSenders) > 0 && !allowedSenders[addr] {
		log.Printf("Rejecting MAIL from non-allowed sender: %s", addr)
		return errors.New("550 5.7.1 contact not permitted")
	}

	s.from = addr
	return nil
}

func (s *session) Rcpt(to string, opts *smtplib.RcptOptions) error {
	addr := normalizeAddress(to)
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
		return s.handleGenerate(r)

	}

	if containsAddr(s.to, "test@soonerfleet.com") {
		return s.handleTest(r)
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
