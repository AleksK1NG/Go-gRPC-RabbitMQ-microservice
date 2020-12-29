package repository

import (
	"context"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/AleksK1NG/email-microservice/pkg/mime_types"
	"github.com/AleksK1NG/email-microservice/pkg/utils"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEmailsRepository_CreateEmail(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	emailsRepository := NewEmailsRepository(sqlxDB)

	emailUUID := uuid.New()
	mockEmail := &models.Email{
		To:          []string{"mail@gmail.com"},
		From:        "alex@gmail.com",
		Body:        "<span>some text content</span>",
		Subject:     "Confirm your email",
		ContentType: mime_types.MIMEApplicationJSON,
	}

	columns := []string{"email_id"}
	rows := sqlmock.NewRows(columns).AddRow(emailUUID)

	mock.ExpectQuery(createEmailQuery).WithArgs(
		mockEmail.GetToString(),
		mockEmail.From,
		mockEmail.Subject,
		mockEmail.Body,
		mockEmail.ContentType,
	).WillReturnRows(rows)

	createdEmail, err := emailsRepository.CreateEmail(context.Background(), mockEmail)
	require.NoError(t, err)
	require.NotNil(t, createdEmail)
	require.Equal(t, mockEmail.To, createdEmail.To)
}

func TestEmailsRepository_FindEmailById(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	emailsRepository := NewEmailsRepository(sqlxDB)

	emailUUID := uuid.New()
	var strTo string
	mockEmail := &models.Email{
		EmailID:     emailUUID,
		To:          []string{"mail@gmail.com"},
		From:        "alex@gmail.com",
		Body:        "<span>some text content</span>",
		Subject:     "Confirm your email",
		ContentType: mime_types.MIMEApplicationJSON,
		CreatedAt:   time.Now(),
	}

	columns := []string{"email_id", "to", "from", "subject", "body", "content_type", "created_at"}
	rows := sqlmock.NewRows(columns).AddRow(
		&mockEmail.EmailID,
		&strTo,
		&mockEmail.From,
		&mockEmail.Body,
		&mockEmail.Subject,
		&mockEmail.ContentType,
		&mockEmail.CreatedAt,
	)

	mock.ExpectQuery(findEmailByIdQuery).WithArgs(emailUUID).WillReturnRows(rows)

	foundEmail, err := emailsRepository.FindEmailById(context.Background(), emailUUID)
	require.NoError(t, err)
	require.NotNil(t, foundEmail)
	require.Equal(t, foundEmail.EmailID, emailUUID)
}

func TestEmailsRepository_FindEmailsByReceiver(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	emailsRepository := NewEmailsRepository(sqlxDB)

	emailUUID := uuid.New()
	var strTo string
	mockEmail := &models.Email{
		EmailID:     emailUUID,
		To:          []string{"mail@gmail.com"},
		From:        "alex@gmail.com",
		Body:        "<span>some text content</span>",
		Subject:     "Confirm your email",
		ContentType: mime_types.MIMEApplicationJSON,
		CreatedAt:   time.Now(),
	}

	pq := &utils.PaginationQuery{
		Size: 10,
		Page: 1,
	}

	columns := []string{"email_id", "to", "from", "subject", "body", "content_type", "created_at"}
	rows := sqlmock.NewRows(columns).AddRow(
		&mockEmail.EmailID,
		&strTo,
		&mockEmail.From,
		&mockEmail.Body,
		&mockEmail.Subject,
		&mockEmail.ContentType,
		&mockEmail.CreatedAt,
	)
	emailListCount := 1
	totalCountColumns := []string{"totalCount"}
	totalCountRows := sqlmock.NewRows(totalCountColumns).AddRow(emailListCount)

	mock.ExpectQuery(totalCountQuery).WithArgs(mockEmail.To[0]).WillReturnRows(totalCountRows)
	mock.ExpectQuery(findEmailByReceiverQuery).WithArgs(mockEmail.To[0], pq.GetOffset(), pq.GetLimit()).WillReturnRows(rows)

	emailsList, err := emailsRepository.FindEmailsByReceiver(context.Background(), mockEmail.To[0], pq)
	require.NoError(t, err)
	require.NotNil(t, emailsList)
	require.Equal(t, len(emailsList.Emails), emailListCount)
}
