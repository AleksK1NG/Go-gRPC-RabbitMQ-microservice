package mailer

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"gopkg.in/gomail.v2"

	"github.com/AleksK1NG/email-microservice/internal/models"
)

// Mailer agent
type Mailer struct {
	mailDialer *gomail.Dialer
}

// NewMailer Mailer agent constructor
func NewMailer(mailDialer *gomail.Dialer) *Mailer {
	return &Mailer{mailDialer: mailDialer}
}

// Send email
func (m *Mailer) Send(ctx context.Context, email *models.Email) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "Mailer.Send")
	defer span.Finish()

	gm := gomail.NewMessage()
	gm.SetHeader("From", email.From)
	gm.SetHeader("To", email.To...)
	gm.SetHeader("Subject", email.Subject)
	gm.SetBody(email.ContentType, email.Body)

	return m.mailDialer.DialAndSend(gm)
}
