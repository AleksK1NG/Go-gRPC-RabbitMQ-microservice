package email

import (
	"context"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/streadway/amqp"
)

// Image useCase interface
type EmailsUseCase interface {
	SendEmail(ctx context.Context, delivery amqp.Delivery) error
	PublishEmailToQueue(ctx context.Context, email *models.Email) error
}
