package domain

type Mailer interface {
	SendVerificationEmail(to, name, token string) error
	SendResetPasswordEmail(to, name, token string) error
	SendCancelRecruitmentEmail(to, name string) error
	SendStopOnboardingEmail(to, name string) error
	SendInterviewReminderEmail(to, name, interviewTitle, scheduledAt, meetingLink string) error
	SendRecruitmentInvitationEmail(to, name, jobRoleName string) error
}
