package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/AleksK1NG/email-microservice/pkg/mime_types"
	"github.com/AleksK1NG/email-microservice/pkg/utils"
)

// Email struct
type Email struct {
	EmailID     uuid.UUID `json:"emailId" db:"email_id" validate:"omitempty"`
	To          []string  `json:"to" db:"to" validate:"required"`
	From        string    `json:"from,omitempty" db:"from" validate:"required,email"`
	Body        string    `json:"body" db:"body" validate:"required"`
	Subject     string    `json:"subject" db:"subject" validate:"required,lte=250"`
	ContentType string    `json:"contentType,omitempty" db:"content_type" validate:"lte=250"`
	CreatedAt   time.Time `json:"created_at,omitempty" db:"created_at"`
}

// Get string from addresses
func (e *Email) GetToString() string {
	return strings.Join(e.To, ",")
}

// Prepare email for creation
func (e *Email) PrepareAndValidate(ctx context.Context) error {
	e.From = strings.TrimSpace(strings.ToLower(e.From))
	for i, mail := range e.To {
		if !utils.ValidateEmail(e.To[i]) {
			return fmt.Errorf("invalid email: %s", mail)
		}
		e.To[i] = strings.TrimSpace(strings.ToLower(e.To[i]))
	}
	e.ContentType = mime_types.MIMEApplicationJSON

	return utils.ValidateStruct(ctx, e)
}

// Set to array from string value
func (e *Email) SetToFromString(to string) {
	e.To = strings.Split(to, ",")
}

// Emails list with pagination
type EmailsList struct {
	TotalCount uint64   `json:"total_count"`
	TotalPages uint64   `json:"total_pages"`
	Page       uint64   `json:"page"`
	Size       uint64   `json:"size"`
	HasMore    bool     `json:"has_more"`
	Emails     []*Email `json:"emails"`
}
