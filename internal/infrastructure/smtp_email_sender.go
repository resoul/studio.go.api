package infrastructure

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	mail "gopkg.in/mail.v2"
)

type smtpEmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
	logoPath string
}

func NewSMTPEmailSender(host string, port int, username, password, from, logoPath string) EmailSender {
	return &smtpEmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		logoPath: logoPath,
	}
}

func (s *smtpEmailSender) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m := mail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	if textBody == "" {
		textBody = subject
	}
	m.SetBody("text/plain", textBody)
	if htmlBody != "" {
		m.AddAlternative("text/html", htmlBody)
		if s.logoPath != "" {
			if _, err := os.Stat(s.logoPath); err == nil {
				m.Embed(s.logoPath, mail.SetHeader(map[string][]string{
					"Content-ID": {"<logo.png>"},
				}))
			} else {
				logrus.WithError(err).WithField("logo_path", s.logoPath).Warn("Mailer logo file not found, sending without logo")
			}
		}
	}

	d := mail.NewDialer(s.host, s.port, s.username, s.password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email via smtp: %w", err)
	}

	return nil
}
