package email

import (
	"context"
	"github.com/AleksK1NG/email-microservice/internal/models"
)

// Mailer interface
type Mailer interface {
	Send(ctx context.Context, email *models.Email) error
}
