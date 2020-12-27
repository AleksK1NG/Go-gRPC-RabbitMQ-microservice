package models

import "github.com/google/uuid"

// Email struct
type Email struct {
	EmailID  uuid.UUID `json:"emailId" db:"email_id" validate:"omitempty"`
	AddrFrom string    `json:"from" db:"addr_from" validate:"required,email"`
	AddrTo   string    `json:"to" db:"addr_to" validate:"required,email"`
	Title    string    `json:"title" db:"title" validate:"required,lte=250"`
	Body     string    `json:"body" db:"body" validate:"required"`
}
