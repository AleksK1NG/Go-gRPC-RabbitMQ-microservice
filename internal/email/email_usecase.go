package email

import (
	"context"
	"github.com/streadway/amqp"
)

// Image useCase interface
type EmailsUseCase interface {
	SendEmail(ctx context.Context, delivery amqp.Delivery) error
}
