package repository

import "github.com/jmoiron/sqlx"

// Images Emails Repository
type EmailsRepository struct {
	db *sqlx.DB
}

// Images AWS repository constructor
func NewEmailsRepository(db *sqlx.DB) *EmailsRepository {
	return &EmailsRepository{db: db}
}

// Send email
func (e *EmailsRepository) SendEmail(email string) (string, error) {
	panic("implement me")
}
