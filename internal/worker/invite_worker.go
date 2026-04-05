package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/resoul/studio.go.api/internal/infrastructure/rabbitmq"
	"github.com/sirupsen/logrus"
)

const (
	inviteQueue    = "workspace.invites"
	inviteExchange = ""
)

// InviteWorker consumes workspace.invites events and dispatches the invite email.
// DB persistence happens in the service before publishing — this worker is
// responsible only for delivery. A failure here does not affect the invite record.
type InviteWorker struct {
	rbmq   *rabbitmq.Client
	repo   domain.WorkspaceRepository
	mailer domain.Mailer
}

func NewInviteWorker(rbmq *rabbitmq.Client, repo domain.WorkspaceRepository, mailer domain.Mailer) *InviteWorker {
	return &InviteWorker{rbmq: rbmq, repo: repo, mailer: mailer}
}

// Start declares the queue and blocks consuming messages until ctx is cancelled.
func (w *InviteWorker) Start(ctx context.Context) {
	if _, err := w.rbmq.DeclareQueue(inviteQueue, true, false, false, false, nil); err != nil {
		logrus.WithError(err).Error("invite worker: failed to declare queue")
		return
	}

	msgs, err := w.rbmq.Consume(inviteQueue, "invite-worker", false, false, false, false, nil)
	if err != nil {
		logrus.WithError(err).Error("invite worker: failed to start consuming")
		return
	}

	logrus.Info("invite worker: listening on queue ", inviteQueue)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("invite worker: shutting down")
			return
		case msg, ok := <-msgs:
			if !ok {
				logrus.Warn("invite worker: channel closed")
				return
			}
			w.handle(ctx, msg)
		}
	}
}

func (w *InviteWorker) handle(ctx context.Context, msg amqp.Delivery) {
	var event domain.InviteEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logrus.WithError(err).Error("invite worker: malformed message, nacking without requeue")
		_ = msg.Nack(false, false) // dead-letter, don't requeue garbage
		return
	}

	if err := w.sendInviteEmail(ctx, event); err != nil {
		logrus.WithError(err).
			WithField("email", event.Email).
			WithField("token", event.Token).
			Warn("invite worker: email delivery failed, requeueing")
		_ = msg.Nack(false, true) // requeue for retry
		return
	}

	_ = msg.Ack(false)
	logrus.WithField("email", event.Email).Info("invite worker: invite email sent")
}

func (w *InviteWorker) sendInviteEmail(ctx context.Context, event domain.InviteEvent) error {
	link := fmt.Sprintf("%s/invites/%s", strings.TrimRight(event.InviteBaseURL, "/"), event.Token)

	expiresAt, _ := time.Parse(time.RFC3339, event.ExpiresAt)
	expiresIn := "7 days"
	if !expiresAt.IsZero() {
		if h := int(time.Until(expiresAt).Hours()); h > 0 {
			expiresIn = fmt.Sprintf("%d hours", h)
		}
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;color:#1a1a1a;max-width:480px;margin:0 auto;padding:32px 16px">
  <h2 style="margin-bottom:8px">You've been invited</h2>
  <p>You have been invited to join <strong>%s</strong> as <strong>%s</strong>.</p>
  <p style="margin:24px 0">
    <a href="%s"
       style="background:#18181b;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-weight:600">
      Accept Invitation
    </a>
  </p>
  <p style="color:#71717a;font-size:13px">This link expires in %s. If you did not expect this invitation, you can ignore this email.</p>
</body>
</html>`, event.WorkspaceName, event.Role, link, expiresIn)

	return w.mailer.Send(ctx, domain.MailMessage{
		To:      []string{event.Email},
		Subject: fmt.Sprintf("You've been invited to %s", event.WorkspaceName),
		HTML:    html,
	})
}
