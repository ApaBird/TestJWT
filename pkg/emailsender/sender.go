package emailsender

import "net/smtp"

type AuthData struct {
	Identity string
	Username string
	Password string
	Host     string
	Email    string
	Addr     string
}

type EmailSender struct {
	auth  smtp.Auth
	Email string
	Addr  string
}

func NewEmailSender(config *AuthData) *EmailSender {
	return &EmailSender{
		auth:  smtp.PlainAuth(config.Identity, config.Username, config.Password, config.Host),
		Email: config.Email,
		Addr:  config.Addr,
	}
}

func (sender *EmailSender) WarningEmail(email string, message string) error {
	err := smtp.SendMail(sender.Addr, sender.auth, sender.Email, []string{email}, []byte(message))
	if err != nil {
		return err
	}

	return nil
}
