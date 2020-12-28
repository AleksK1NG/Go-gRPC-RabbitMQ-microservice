package usecase

import (
	"context"
	"encoding/json"
	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/internal/email"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/AleksK1NG/email-microservice/pkg/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Image useCase
type EmailUseCase struct {
	mailer     email.Mailer
	emailsRepo email.EmailsRepository
	logger     logger.Logger
	cfg        *config.Config
}

// Image useCase constructor
func NewEmailUseCase(emailsRepo email.EmailsRepository, logger logger.Logger, mailer email.Mailer, cfg *config.Config) *EmailUseCase {
	return &EmailUseCase{emailsRepo: emailsRepo, logger: logger, mailer: mailer, cfg: cfg}
}

// Send email
func (e *EmailUseCase) SendEmail(ctx context.Context, delivery amqp.Delivery) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.SendEmail")
	defer span.Finish()

	mail := &models.Email{}
	if err := json.Unmarshal(delivery.Body, mail); err != nil {
		return errors.Wrap(err, "json.Unmarshal")
	}

	mail.From = e.cfg.Smtp.User
	if err := utils.ValidateStruct(ctx, mail); err != nil {
		return errors.Wrap(err, "ValidateStruct")
	}

	//e.logger.Infof("SendEmail: %#v", mail)
	//if err := e.mailer.Send(ctx, mail); err != nil {
	//	return errors.Wrap(err, "mailer.Send")
	//}

	createdEmail, err := e.emailsRepo.CreateEmail(ctx, mail)
	if err != nil {
		return errors.Wrap(err, "emailsRepo.CreateEmail")
	}

	span.LogFields(log.String("emailID", createdEmail.EmailID.String()))
	e.logger.Infof("Success send email: %v", createdEmail.EmailID)
	return nil
}
