package email

import (
	"context"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/google/uuid"
)

// Image Repository interface
type EmailsRepository interface {
	CreateEmail(ctx context.Context, email *models.Email) (*models.Email, error)
	FindEmailById(ctx context.Context, emailID uuid.UUID) (*models.Email, error)
}
