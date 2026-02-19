package infrastructure

import (
	"context"

	"github.com/sirupsen/logrus"
)

type logEmailSender struct{}

func NewLogEmailSender() EmailSender {
	return &logEmailSender{}
}

func (s *logEmailSender) Send(_ context.Context, to, subject, body string) error {
	logrus.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
		"body":    body,
	}).Info("Email queued (log sender)")

	return nil
}
