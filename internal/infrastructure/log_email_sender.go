package infrastructure

import (
	"context"

	"github.com/sirupsen/logrus"
)

type logEmailSender struct{}

func NewLogEmailSender() EmailSender {
	return &logEmailSender{}
}

func (s *logEmailSender) Send(_ context.Context, to, subject, textBody, htmlBody string) error {
	logrus.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
		"text":    textBody,
		"html":    htmlBody,
	}).Info("Email queued (log sender)")

	return nil
}
