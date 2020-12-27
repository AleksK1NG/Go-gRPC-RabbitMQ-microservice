package mailer

import (
	"context"
	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/opentracing/opentracing-go"
	"gopkg.in/gomail.v2"
)

// Mailer agent
type Mailer struct {
	cfg        *config.Config
	mailDialer *gomail.Dialer
}

// Mailer agent constructor
func NewMailer(cfg *config.Config, mailDialer *gomail.Dialer) *Mailer {
	return &Mailer{cfg: cfg, mailDialer: mailDialer}
}

// Send email
func (m *Mailer) Send(ctx context.Context, email *models.Email) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Mailer.Send")
	defer span.Finish()

	gm := gomail.NewMessage()
	gm.SetHeader("From", email.From)
	gm.SetHeader("To", email.To...)
	gm.SetHeader("Subject", email.Subject)
	gm.SetBody(email.ContentType, email.Body)

	return m.mailDialer.DialAndSend(gm)
}
