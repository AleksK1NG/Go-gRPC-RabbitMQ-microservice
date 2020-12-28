package repository

import (
	"context"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/AleksK1NG/email-microservice/pkg/utils"
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

// Find email by id
func (e *EmailsRepository) FindEmailById(ctx context.Context, emailID uuid.UUID) (*models.Email, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailsRepository.FindEmailById")
	defer span.Finish()

	findEmailByIdQuery := `SELECT email_id, "to", "from", subject, body, content_type, created_at FROM emails WHERE email_id = $1`

	var to string
	email := &models.Email{}
	if err := e.db.QueryRowContext(ctx, findEmailByIdQuery, emailID).Scan(
		&email.EmailID,
		&to,
		&email.From,
		&email.Subject,
		&email.Body,
		&email.ContentType,
		&email.CreatedAt,
	); err != nil {
		return nil, errors.Wrap(err, "db.QueryRowContext")
	}
	email.SetToFromString(to)

	return email, nil
}

// Find emails by receiver
func (e *EmailsRepository) FindEmailsByReceiver(ctx context.Context, to string, query *utils.PaginationQuery) (*models.EmailsList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailsRepository.FindEmailsByReceiver")
	defer span.Finish()

	totalCountQuery := `SELECT COUNT (email_id) as totalCount FROM emails WHERE "to" ILIKE '%' || $1 || '%'`
	var totalCount uint64
	if err := e.db.QueryRowContext(ctx, totalCountQuery, to).Scan(&totalCount); err != nil {
		return nil, errors.Wrap(err, "db.QueryRowContext")
	}
	if totalCount == 0 {
		return &models.EmailsList{Emails: []*models.Email{}}, nil
	}

	findEmailByReceiverQuery := `SELECT email_id, "to", "from", subject, body, content_type, created_at 
	FROM emails WHERE "to" ILIKE '%' || $1 || '%' ORDER BY created_at OFFSET $2 LIMIT $3`

	rows, err := e.db.QueryxContext(ctx, findEmailByReceiverQuery, to, query.GetOffset(), query.GetLimit())
	if err != nil {
		return nil, errors.Wrap(err, "db.QueryxContext")
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	emails := make([]*models.Email, 0, query.GetSize())
	for rows.Next() {
		var mailTo string
		email := &models.Email{}
		if err := rows.Scan(
			&email.EmailID,
			&mailTo,
			&email.From,
			&email.Subject,
			&email.Body,
			&email.ContentType,
			&email.CreatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		email.SetToFromString(mailTo)
		emails = append(emails, email)
	}

	return &models.EmailsList{
		TotalCount: totalCount,
		TotalPages: utils.GetTotalPages(totalCount, query.GetSize()),
		Page:       query.GetPage(),
		Size:       query.GetSize(),
		HasMore:    utils.GetHasMore(query.GetPage(), totalCount, query.GetSize()),
		Emails:     emails,
	}, err
}
