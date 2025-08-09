package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"time"

	emails "github.com/elliot40404/mailc/examples/generated"
)

// sendSMTP sends an email using net/smtp with STARTTLS when possible.
func sendSMTP(host string, port int, username, password, from, to, subject, html string) error {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	// Establish TCP connection
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = c.Quit() }()

	// Upgrade to TLS if supported
	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	// Authenticate if supported
	if ok, _ := c.Extension("AUTH"); ok && username != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}

	if err := c.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}

	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	defer wc.Close()

	// Minimal MIME headers
	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" + html

	if _, err := wc.Write([]byte(msg)); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

func main() {
	// Render an example email from examples/generated
	data := &emails.WelcomePersonalizedEmailData{
		Username:  "jane@example.com",
		FirstName: "Jane",
	}
	res, err := emails.WelcomePersonalizedEmail(data)
	if err != nil {
		log.Fatalf("render: %v", err)
	}

	// SMTP settings from environment
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	to := os.Getenv("SMTP_TO")

	if host == "" || portStr == "" || from == "" || to == "" {
		log.Printf("SMTP not configured; printing output instead\nSubject: %s\n\n%s\n", res.Subject, res.HTML)
		return
	}
	port, _ := strconv.Atoi(portStr)

	if err := sendSMTP(host, port, user, pass, from, to, res.Subject, res.HTML); err != nil {
		log.Fatalf("send: %v", err)
	}

	log.Println("email sent")
}
