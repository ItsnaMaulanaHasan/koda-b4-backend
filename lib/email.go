package lib

import (
	"os"

	resend "github.com/resend/resend-go/v2"
)

func SendResendEmail(toEmail, subject, htmlBody string) error {
	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))

	params := &resend.SendEmailRequest{
		From:    os.Getenv("EMAIL_FROM"),
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
	}

	_, err := client.Emails.Send(params)
	return err
}
