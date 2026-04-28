package mailgunrepo

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/mailgun/mailgun-go/v4"
)

type mailgunRepo struct {
	mg        *mailgun.MailgunImpl
	from      string
	fromName  string
	appURL    string
}

func NewMailgunRepo() domain.Mailer {
	domainName := os.Getenv("MAILGUN_DOMAIN")
	apiKey := os.Getenv("MAILGUN_API_KEY")
	fromEmail := os.Getenv("MAILGUN_FROM_EMAIL")
	fromName := os.Getenv("MAILGUN_FROM_NAME")
	appURL := os.Getenv("FRONTEND_URL")

	mg := mailgun.NewMailgun(domainName, apiKey)

	return &mailgunRepo{
		mg:       mg,
		from:     fromEmail,
		fromName: fromName,
		appURL:   appURL,
	}
}

func (r *mailgunRepo) SendVerificationEmail(to, name, token string) error {
	subject := "Verify Your Account - GigSourceHub"
	verifyURL := fmt.Sprintf("%s/verify?token=%s", r.appURL, token)

	body := fmt.Sprintf(`
		Hi %s,

		Welcome to GigSourceHub! Please verify your account by clicking the link below:
		
		%s

		This link will expire in 24 hours.

		Best regards,
		GigSourceHub Team
	`, name, verifyURL)

	mgMessage := r.mg.NewMessage(fmt.Sprintf("%s <%s>", r.fromName, r.from), subject, body, to)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err := r.mg.Send(ctx, mgMessage)
	return err
}

func (r *mailgunRepo) SendResetPasswordEmail(to, name, token string) error {
	subject := "Reset Your Password - GigSourceHub"
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", r.appURL, token)

	body := fmt.Sprintf(`
		Hi %s,

		We received a request to reset your password. You can do so by clicking the link below:

		%s

		This link will expire in 1 hour.

		If you did not request a password reset, please ignore this email.

		Best regards,
		GigSourceHub Team
	`, name, resetURL)

	mgMessage := r.mg.NewMessage(fmt.Sprintf("%s <%s>", r.fromName, r.from), subject, body, to)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err := r.mg.Send(ctx, mgMessage)
	return err
}
