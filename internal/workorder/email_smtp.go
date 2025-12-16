package workorder

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"mime"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type SMTPSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

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
	return &SMTPSender{host: host, port: port, username: username, password: password, from: from}, nil
}

func NewSMTPSenderFromEnv() (*SMTPSender, error) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}

	port := 587
	if s := os.Getenv("SMTP_PORT"); s != "" {
		p, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid SMTP_PORT %q: %w", s, err)
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

// SendWorkOrderReport sends a plain text email with a PDF attachment.
func (s *SMTPSender) SendWorkOrderReport(
	ctx context.Context,
	to string,
	customerName string,
	workOrderCode string,
	pdfBytes []byte,
) error {
	subject := fmt.Sprintf("HavenzSure Report – %s", workOrderCode)

	body := fmt.Sprintf(`Hi %s,

Please find attached your work order report (%s).

Thank you,
HavenzSure Team
`, customerName, workOrderCode)

	filename := workOrderCode + ".pdf"
	return s.sendWithPdfAttachment(to, subject, body, filename, pdfBytes)
}

func (s *SMTPSender) sendWithPdfAttachment(to, subject, body, filename string, pdf []byte) error {
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("smtp: recipient email is empty")
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	boundary := fmt.Sprintf("mixed-%d", time.Now().UnixNano())

	encodedFilename := mime.QEncoding.Encode("UTF-8", filename)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%q\r\n", boundary))
	sb.WriteString("\r\n")

	// ---- text part
	sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	sb.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	sb.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		sb.WriteString("\r\n")
	}
	sb.WriteString("\r\n")

	// ---- pdf attachment part
	sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	sb.WriteString("Content-Type: application/pdf\r\n")
	sb.WriteString("Content-Transfer-Encoding: base64\r\n")
	sb.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", encodedFilename))
	sb.WriteString("\r\n")

	// base64 wrap at 76 chars/line
	b64 := base64.StdEncoding.EncodeToString(pdf)
	for i := 0; i < len(b64); i += 76 {
		end := i + 76
		if end > len(b64) {
			end = len(b64)
		}
		sb.WriteString(b64[i:end])
		sb.WriteString("\r\n")
	}

	sb.WriteString("\r\n")
	sb.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	msg := []byte(sb.String())

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, msg); err != nil {
		return fmt.Errorf("smtp send to %s failed: %w", to, err)
	}

	log.Printf("SMTP report email sent to %s with subject %q (attachment=%s)", to, subject, filename)
	return nil
}
