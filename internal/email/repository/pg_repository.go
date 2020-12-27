package repository

// Images Emails Repository
type EmailsRepository struct{}

// Images AWS repository constructor
func NewEmailsRepository() *EmailsRepository {
	return &EmailsRepository{}
}

// Send email
func (e *EmailsRepository) SendEmail(email string) (string, error) {
	panic("implement me")
}
