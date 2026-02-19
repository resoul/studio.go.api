package infrastructure

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

type smtpEmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPEmailSender(host string, port int, username, password, from string) EmailSender {
	return &smtpEmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *smtpEmailSender) Send(_ context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", s.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if s.username != "" || s.password != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send email via smtp: %w", err)
	}

	return nil
}
