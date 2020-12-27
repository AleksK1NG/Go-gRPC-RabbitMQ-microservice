package usecase

import (
	"context"
	"encoding/json"
	"github.com/AleksK1NG/email-microservice/internal/email"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Image useCase
type EmailUseCase struct {
	mailer     email.Mailer
	emailsRepo email.EmailsRepository
	logger     logger.Logger
}

// Image useCase constructor
func NewEmailUseCase(emailsRepo email.EmailsRepository, logger logger.Logger, mailer email.Mailer) *EmailUseCase {
	return &EmailUseCase{emailsRepo: emailsRepo, logger: logger, mailer: mailer}
}

// Send email
func (e *EmailUseCase) SendEmail(ctx context.Context, delivery amqp.Delivery) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.SendEmail")
	defer span.Finish()

	mail := &models.Email{}
	if err := json.Unmarshal(delivery.Body, mail); err != nil {
		return errors.Wrap(err, "json.Unmarshal")
	}

	e.logger.Infof("SendEmail: %#v", mail)

	return nil
}
