package usecase

import (
	"context"
	"encoding/json"
	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/internal/email"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/AleksK1NG/email-microservice/pkg/utils"
	"github.com/google/uuid"
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
	publisher  email.EmailsPublisher
}

// Image useCase constructor
func NewEmailUseCase(emailsRepo email.EmailsRepository, logger logger.Logger, mailer email.Mailer, cfg *config.Config, publisher email.EmailsPublisher) *EmailUseCase {
	return &EmailUseCase{emailsRepo: emailsRepo, logger: logger, mailer: mailer, cfg: cfg, publisher: publisher}
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

// Publish email to rabbitmq
func (e *EmailUseCase) PublishEmailToQueue(ctx context.Context, email *models.Email) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.PublishEmailToQueue")
	defer span.Finish()

	mailBytes, err := json.Marshal(email)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	return e.publisher.Publish(mailBytes, email.ContentType)
}

// Find email by uuid
func (e *EmailUseCase) FindEmailById(ctx context.Context, emailID uuid.UUID) (*models.Email, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.FindEmailById")
	defer span.Finish()

	return e.emailsRepo.FindEmailById(ctx, emailID)
}

// Find emails by receiver
func (e *EmailUseCase) FindEmailsByReceiver(ctx context.Context, to string, query *utils.PaginationQuery) (*models.EmailsList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.SendEmail")
	defer span.Finish()

	return e.emailsRepo.FindEmailsByReceiver(ctx, to, query)
}
