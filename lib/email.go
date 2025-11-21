package lib

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	AppUrl       string
}

func SendPasswordResetEmail(toEmail, token string) error {
	config := EmailConfig{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
		AppUrl:       os.Getenv("APP_URL"),
	}

	if config.SMTPHost == "" || config.SMTPPort == "" || config.SMTPUsername == "" || config.SMTPPassword == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)

	resetLink := fmt.Sprintf("%s/reset-password?email=%s&token=%s", config.AppUrl, toEmail, token)

	subject := "Subject: Password Reset Request\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	body := fmt.Sprintf(`
		<html>
			<body>
				<h2>Password Reset Request</h2>
				<p>You have requested to reset your password. Click the link below to reset your password:</p>
				<p><a href="%s">Reset Password</a></p>
				<p>Or copy and paste this link into your browser:</p>
				<p>%s</p>
				<p>This link will expire in 1 hour.</p>
				<p>If you did not request this, please ignore this email.</p>
			</body>
		</html>
	`, resetLink, resetLink)

	message := []byte(subject + mime + body)

	addr := fmt.Sprintf("%s:%s", config.SMTPHost, config.SMTPPort)
	err := smtp.SendMail(addr, auth, config.FromEmail, []string{toEmail}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
