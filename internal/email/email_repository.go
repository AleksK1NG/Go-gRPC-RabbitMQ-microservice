package email

import (
	"context"
	"github.com/AleksK1NG/email-microservice/internal/models"
)

// Image Repository interface
type EmailsRepository interface {
	CreateEmail(ctx context.Context, email *models.Email) (*models.Email, error)
}
