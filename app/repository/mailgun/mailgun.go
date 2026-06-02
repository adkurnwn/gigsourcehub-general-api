package mailgunrepo

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/mailgun/mailgun-go/v4"
)

//go:embed templates/*.tmpl
var emailTemplatesFS embed.FS

var emailTemplates = template.Must(template.New("emails").ParseFS(emailTemplatesFS, "templates/*.tmpl"))

type mailgunRepo struct {
	mg       *mailgun.MailgunImpl
	from     string
	fromName string
	appURL   string
	logoURL  string
}

type emailTemplateData struct {
	Subject    string
	AppURL     string
	Name       string
	ActionURL  string
	ButtonText string
	LogoURL    string
}

func NewMailgunRepo() domain.Mailer {
	domainName := os.Getenv("MAILGUN_DOMAIN")
	apiKey := os.Getenv("MAILGUN_API_KEY")
	fromEmail := os.Getenv("MAILGUN_FROM_EMAIL")
	fromName := os.Getenv("MAILGUN_FROM_NAME")
	appURL := os.Getenv("FRONTEND_URL")
	logoURL := os.Getenv("EMAIL_LOGO_URL")
	if logoURL == "" {
		logoURL = "https://cdn.magangslab.store/assets/gigsourcehub-logo.png"
	}

	mg := mailgun.NewMailgun(domainName, apiKey)

	return &mailgunRepo{
		mg:       mg,
		from:     fromEmail,
		fromName: fromName,
		appURL:   appURL,
		logoURL:  logoURL,
	}
}

func (r *mailgunRepo) renderHTMLTemplate(templateName string, data emailTemplateData) (string, error) {
	var buf bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (r *mailgunRepo) makeEmailTemplateData(subject, name, actionURL, buttonText string) emailTemplateData {
	return emailTemplateData{
		Subject:    subject,
		AppURL:     r.appURL,
		Name:       name,
		ActionURL:  actionURL,
		ButtonText: buttonText,
		LogoURL:    r.logoURL,
	}
}

func (r *mailgunRepo) SendVerificationEmail(to, name, token string) error {
	subject := "Verify Your Account - GigSourceHub"
	verifyURL := fmt.Sprintf("%s/verify?token=%s", r.appURL, token)

	data := r.makeEmailTemplateData(subject, name, verifyURL, "Verify Account")

	htmlBody, err := r.renderHTMLTemplate("verification", data)
	if err != nil {
		return err
	}

	plainBody := fmt.Sprintf(
		"Hi %s,\n\nWelcome to GigSourceHub! Please verify your account by visiting the link below:\n\n%s\n\nThis link will expire in 24 hours.\n\nBest regards,\nGigSourceHub Team",
		name,
		verifyURL,
	)

	mgMessage := r.mg.NewMessage(fmt.Sprintf("%s <%s>", r.fromName, r.from), subject, plainBody, to)
	mgMessage.SetHtml(htmlBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err = r.mg.Send(ctx, mgMessage)
	return err
}

func (r *mailgunRepo) SendResetPasswordEmail(to, name, token string) error {
	subject := "Reset Your Password - GigSourceHub"
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", r.appURL, token)

	data := r.makeEmailTemplateData(subject, name, resetURL, "Reset Password")

	htmlBody, err := r.renderHTMLTemplate("reset_password", data)
	if err != nil {
		return err
	}

	plainBody := fmt.Sprintf(
		"Hi %s,\n\nWe received a request to reset your password. Use the link below to continue:\n\n%s\n\nThis link will expire in 1 hour. If you did not request this, please ignore this email.\n\nBest regards,\nGigSourceHub Team",
		name,
		resetURL,
	)

	mgMessage := r.mg.NewMessage(fmt.Sprintf("%s <%s>", r.fromName, r.from), subject, plainBody, to)
	mgMessage.SetHtml(htmlBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err = r.mg.Send(ctx, mgMessage)
	return err
}

func (r *mailgunRepo) SendCancelRecruitmentEmail(to, name string) error {
	subject := "Update on Your Recruitment Process - GigSourceHub"

	data := r.makeEmailTemplateData(subject, name, "", "")

	htmlBody, err := r.renderHTMLTemplate("cancel_recruitment", data)
	if err != nil {
		return err
	}

	plainBody := fmt.Sprintf(
		"Hi %s,\n\nWe wanted to update you on your recruitment process. Unfortunately, your application will not be moving forward at this time.\n\nPlease note that your chat history with our HR team will be automatically deleted in 12 hours.\n\nThank you for your interest and time.\n\nBest regards,\nGigSourceHub Team",
		name,
	)

	mgMessage := r.mg.NewMessage(fmt.Sprintf("%s <%s>", r.fromName, r.from), subject, plainBody, to)
	mgMessage.SetHtml(htmlBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err = r.mg.Send(ctx, mgMessage)
	return err
}

func (r *mailgunRepo) SendStopOnboardingEmail(to, name string) error {
	subject := "Update on Your Onboarding Contract - GigSourceHub"

	data := r.makeEmailTemplateData(subject, name, "", "")

	htmlBody, err := r.renderHTMLTemplate("stop_onboarding", data)
	if err != nil {
		return err
	}

	plainBody := fmt.Sprintf(
		"Hi %s,\n\nWe are sorry, but we have to cancel your onboarding contract at this time. Your onboarding record has been stopped and any related HR process will be updated accordingly.\n\nIf you need clarification, please reply to this email.\n\nBest regards,\nGigSourceHub Team",
		name,
	)

	mgMessage := r.mg.NewMessage(fmt.Sprintf("%s <%s>", r.fromName, r.from), subject, plainBody, to)
	mgMessage.SetHtml(htmlBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err = r.mg.Send(ctx, mgMessage)
	return err
}
