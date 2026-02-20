package mailer

import "context"

type EmailSender interface {
	Send(ctx context.Context, to, subject, textBody, htmlBody string) error
}
