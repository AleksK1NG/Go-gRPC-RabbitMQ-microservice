//go:generate mockgen -source pg_repository.go -destination mock/pg_repository.go -package mock
package email

import (
	"context"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/AleksK1NG/email-microservice/pkg/utils"
	"github.com/google/uuid"
)

// Image Repository interface
type EmailsRepository interface {
	CreateEmail(ctx context.Context, email *models.Email) (*models.Email, error)
	FindEmailById(ctx context.Context, emailID uuid.UUID) (*models.Email, error)
	FindEmailsByReceiver(ctx context.Context, to string, query *utils.PaginationQuery) (*models.EmailsList, error)
}
