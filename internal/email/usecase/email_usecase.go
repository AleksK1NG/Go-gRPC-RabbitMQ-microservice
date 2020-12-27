package usecase

import (
	"bytes"
	"context"
	"github.com/AleksK1NG/email-microservice/internal/email"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"io/ioutil"
)

// Image useCase
type EmailUseCase struct {
	emailsRepo email.EmailsRepository
	logger     logger.Logger
}

// Image useCase constructor
func NewEmailUseCase(emailsRepo email.EmailsRepository, logger logger.Logger) *EmailUseCase {
	return &EmailUseCase{emailsRepo: emailsRepo, logger: logger}
}

// Send email
func (e *EmailUseCase) SendEmail(ctx context.Context, delivery amqp.Delivery) error {
	reader := bytes.NewReader(delivery.Body)
	deliveryBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadAll")
	}

	e.logger.Infof("SendEmail: %s", string(deliveryBytes))

	return nil
}
