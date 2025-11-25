// Source note:
// This file was partially generated / refactored with assistance from AI (ChatGPT).
// Edited and reviewed by AN-NI HUANG
// Date: 2025-11-22
package user

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

// SMTPSender is a generic SMTP-based implementation of EmailSender.
// For now you'll use it with Gmail's SMTP (smtp.gmail.com:587).
type SMTPSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// NewSMTPSender creates a new SMTPSender with explicit configuration.
func NewSMTPSender(host string, port int, username, password, from string) (*SMTPSender, error) {
	if host == "" {
		return nil, fmt.Errorf("smtp: host must be provided")
	}
	if username == "" || password == "" {
		return nil, fmt.Errorf("smtp: username and password must be provided")
	}
	if from == "" {
		from = username
	}

	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}, nil
}

// NewSMTPSenderFromEnv creates a SMTPSender using environment variables.
// This is convenient for local dev (e.g. Gmail App Password).
//
//   - SMTP_HOST      (default: smtp.gmail.com)
//   - SMTP_PORT      (default: 587)
//   - SMTP_USERNAME  (required, e.g. your Gmail address)
//   - SMTP_PASSWORD  (required, e.g. Gmail App Password)
//   - SMTP_FROM      (optional; if empty, defaults to SMTP_USERNAME)
func NewSMTPSenderFromEnv() (*SMTPSender, error) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}

	port := 587
	if portStr := os.Getenv("SMTP_PORT"); portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SMTP_PORT %q: %w", portStr, err)
		}
		port = p
	}

	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = username
	}

	return NewSMTPSender(host, port, username, password, from)
}

// SendWelcomePasswordSetup implements EmailSender for the welcome flow.
func (s *SMTPSender) SendWelcomePasswordSetup(ctx context.Context, email, firstName, link string) error {
	subject := "Welcome to HavenzSure – Set your password"

	body := fmt.Sprintf(`Hi %s,

Your HavenzSure account has been created successfully.

To complete your account setup, please click the link below to create your password:

%s

⚠️  This link is valid for 1 hour.

If you have any questions, please contact your administrator.

Best regards,
Havenz Tech Team
`, firstName, link)

	return s.sendPlainText(email, subject, body)
}

// SendPasswordSetupReminder implements EmailSender for the reminder flow.
func (s *SMTPSender) SendPasswordSetupReminder(ctx context.Context, email, firstName, link string) error {
	subject := "Reminder – Set your HavenzSure password"

	body := fmt.Sprintf(`Hi %s,

This is a friendly reminder to complete your HavenzSure account setup.

Please use the following link to set your password:

%s

⚠️  This link is valid for 1 hour.

If you have any questions, please contact your administrator.

Best regards,
Havenz Tech Team
`, firstName, link)

	return s.sendPlainText(email, subject, body)
}

// sendPlainText builds a standard plaintext email with basic headers and sends it via SMTP.
func (s *SMTPSender) sendPlainText(to, subject, body string) error {
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("smtp: recipient email is empty")
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)

	msg := []byte(sb.String())

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, msg); err != nil {
		return fmt.Errorf("smtp send to %s failed: %w", to, err)
	}

	log.Printf("SMTP email sent to %s with subject %q", to, subject)
	return nil
}
