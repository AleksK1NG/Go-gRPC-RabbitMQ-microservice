package models

import "github.com/google/uuid"

// Email struct
type Email struct {
	EmailID     uuid.UUID `json:"emailId" db:"email_id" validate:"omitempty"`
	To          []string  `json:"to" db:"to" validate:"required"`
	From        string    `json:"from" db:"from" validate:"required,email"`
	Body        string    `json:"body" db:"body" validate:"required"`
	Subject     string    `json:"subject" db:"subject" validate:"required,lte=250"`
	ContentType string    `json:"contentType" db:"content_type" validate:"required,lte=250"`
}
