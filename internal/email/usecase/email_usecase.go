package usecase

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/internal/email"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/AleksK1NG/email-microservice/pkg/utils"
)

// EmailUseCase Image useCase
type EmailUseCase struct {
	mailer     email.Mailer
	emailsRepo email.EmailsRepository
	logger     logger.Logger
	cfg        *config.Config
	publisher  email.EmailsPublisher
}

// NewEmailUseCase Image useCase constructor
func NewEmailUseCase(emailsRepo email.EmailsRepository, logger logger.Logger, mailer email.Mailer, cfg *config.Config, publisher email.EmailsPublisher) *EmailUseCase {
	return &EmailUseCase{emailsRepo: emailsRepo, logger: logger, mailer: mailer, cfg: cfg, publisher: publisher}
}

// SendEmail Send email
func (e *EmailUseCase) SendEmail(ctx context.Context, deliveryBody []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.SendEmail")
	defer span.Finish()

	mail := &models.Email{}
	if err := json.Unmarshal(deliveryBody, mail); err != nil {
		return errors.Wrap(err, "json.Unmarshal")
	}

	mail.Body = utils.SanitizeString(mail.Body)

	mail.From = e.cfg.Smtp.User
	if err := utils.ValidateStruct(ctx, mail); err != nil {
		return errors.Wrap(err, "ValidateStruct")
	}

	if err := e.mailer.Send(ctx, mail); err != nil {
		return errors.Wrap(err, "mailer.Send")
	}

	createdEmail, err := e.emailsRepo.CreateEmail(ctx, mail)
	if err != nil {
		return errors.Wrap(err, "emailsRepo.CreateEmail")
	}

	span.LogFields(log.String("emailID", createdEmail.EmailID.String()))
	e.logger.Infof("Success send email: %v", createdEmail.EmailID)
	return nil
}

// PublishEmailToQueue Publish email to rabbitmq
func (e *EmailUseCase) PublishEmailToQueue(ctx context.Context, email *models.Email) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "EmailUseCase.PublishEmailToQueue")
	defer span.Finish()

	mailBytes, err := json.Marshal(email)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	return e.publisher.Publish(mailBytes, email.ContentType)
}

// FindEmailById Find email by uuid
func (e *EmailUseCase) FindEmailById(ctx context.Context, emailID uuid.UUID) (*models.Email, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.FindEmailById")
	defer span.Finish()

	return e.emailsRepo.FindEmailById(ctx, emailID)
}

// FindEmailsByReceiver Find emails by receiver
func (e *EmailUseCase) FindEmailsByReceiver(ctx context.Context, to string, query *utils.PaginationQuery) (*models.EmailsList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.FindEmailsByReceiver")
	defer span.Finish()

	return e.emailsRepo.FindEmailsByReceiver(ctx, to, query)
}
