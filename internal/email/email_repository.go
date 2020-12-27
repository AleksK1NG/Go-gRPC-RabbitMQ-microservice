package email

// Image Repository interface
type EmailsRepository interface {
	SendEmail(email string) (string, error)
}
