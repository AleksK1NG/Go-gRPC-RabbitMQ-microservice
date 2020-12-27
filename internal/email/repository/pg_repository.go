package repository

import (
	"context"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// Images Emails Repository
type EmailsRepository struct {
	db *sqlx.DB
}

// Images AWS repository constructor
func NewEmailsRepository(db *sqlx.DB) *EmailsRepository {
	return &EmailsRepository{db: db}
}

// Create email
func (e *EmailsRepository) CreateEmail(ctx context.Context, email *models.Email) (*models.Email, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailsRepository.CreateEmail")
	defer span.Finish()

	createEmailQuery := `INSERT INTO emails ("to", "from", subject, body, content_type) VALUES ($1, $2, $3, $4, $5) RETURNING email_id`

	var id uuid.UUID
	if err := e.db.QueryRowxContext(ctx, createEmailQuery, email.GetToString(), email.From, email.Subject, email.Body, email.ContentType).Scan(&id); err != nil {
		return nil, errors.Wrap(err, "db.QueryRowxContext")
	}

	email.EmailID = id
	return email, nil
}
