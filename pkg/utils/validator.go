package utils

import (
	"context"
	"github.com/go-playground/validator/v10"
	"regexp"
)

// Use a single instance of Validate, it caches struct info
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate struct fields
func ValidateStruct(ctx context.Context, s interface{}) error {
	return validate.StructCtx(ctx, s)
}

// Validate email address
func ValidateEmail(email string) bool {
	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	return emailRegex.MatchString(email)
}
