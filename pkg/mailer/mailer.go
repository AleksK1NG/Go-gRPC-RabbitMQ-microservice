package mailer

import (
	"github.com/AleksK1NG/email-microservice/config"
	"gopkg.in/gomail.v2"
)

// New Mail dialer
func NewMailDialer(cfg *config.Config) *gomail.Dialer {
	return gomail.NewDialer(cfg.Smtp.Host, cfg.Smtp.Port, cfg.Smtp.User, cfg.Smtp.Password)
}
