package domain

type Mailer interface {
	SendVerificationEmail(to, name, token string) error
	SendResetPasswordEmail(to, name, token string) error
	SendCancelRecruitmentEmail(to, name string) error
	SendStopOnboardingEmail(to, name string) error
}
