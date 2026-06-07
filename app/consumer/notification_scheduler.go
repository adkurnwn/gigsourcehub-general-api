package consumer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/sirupsen/logrus"
)

// NotificationScheduler is a background worker that fires time-based notifications.
// It uses a ticker pattern consistent with the existing codebase (no external cron library).
type NotificationScheduler struct {
	repo   domain.GormRepo
	mailer domain.Mailer
}

func NewNotificationScheduler(repo domain.GormRepo, mailer domain.Mailer) *NotificationScheduler {
	return &NotificationScheduler{repo: repo, mailer: mailer}
}

// Start launches the scheduler. It blocks until ctx is cancelled.
// Run in a goroutine: go scheduler.Start(ctx)
func (s *NotificationScheduler) Start(ctx context.Context) {
	logrus.Info("NotificationScheduler: starting background scheduler")
	// Ticker fires every 15 minutes for interview reminders
	interviewTicker := time.NewTicker(15 * time.Minute)
	// Ticker fires every 24 hours for contract expiry check
	contractTicker := time.NewTicker(24 * time.Hour)

	defer interviewTicker.Stop()
	defer contractTicker.Stop()

	// Run once immediately on startup so we don't have to wait for the first tick
	s.checkInterviewReminders(ctx)
	s.checkExpiringContracts(ctx)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("NotificationScheduler: context cancelled, stopping")
			return
		case <-interviewTicker.C:
			s.checkInterviewReminders(ctx)
		case <-contractTicker.C:
			s.checkExpiringContracts(ctx)
		}
	}
}

// checkInterviewReminders sends:
//   - 1-day reminder to candidates (interviews scheduled in the next 24 hours)
//   - 1-hour reminder to admins (interviews scheduled in the next 1 hour)
func (s *NotificationScheduler) checkInterviewReminders(ctx context.Context) {
	logrus.Info("NotificationScheduler: checking interview reminders...")
	now := time.Now()

	// --- Candidate: 1 day (24h) before ---
	candidateTo := now.Add(24 * time.Hour)
	candidateInterviews, err := s.repo.GetUpcomingInterviewsFor24hReminder(ctx, now, candidateTo)
	if err != nil {
		logrus.Errorf("NotificationScheduler: error fetching candidate interview reminders: %v", err)
	} else {
		logrus.Infof("NotificationScheduler: candidate checks finished, found %d upcoming interviews", len(candidateInterviews))
		for _, iv := range candidateInterviews {
			if iv.CandidateUserID == "" {
				continue
			}
			stageName := ""
			if iv.Stage != nil {
				stageName = iv.Stage.Name
			}
			title := "Interview Reminder"
			desc := fmt.Sprintf("You have an interview scheduled tomorrow. Stage: %s", stageName)
			if iv.Title != "" {
				desc = fmt.Sprintf("You have an interview scheduled tomorrow: %s (Stage: %s)", iv.Title, stageName)
			}
			helpers.SendNotificationAsync(ctx, s.repo, iv.CandidateUserID, title, desc)

			// Send email reminder
			if iv.CandidateUser != nil && iv.CandidateUser.Email != "" {
				candidateName := iv.CandidateUser.Name
				candidateEmail := iv.CandidateUser.Email
				interviewTitle := iv.Title
				if interviewTitle == "" {
					interviewTitle = fmt.Sprintf("Stage: %s", stageName)
				}
				scheduledTimeStr := iv.ScheduledAt.Format("Monday, 02 Jan 2006 at 15:04 MST")
				meetingLink := ""
				if iv.MeetingLink != nil {
					meetingLink = *iv.MeetingLink
				}

				go func(to, name, title, scheduledAt, link string) {
					if err := s.mailer.SendInterviewReminderEmail(to, name, title, scheduledAt, link); err != nil {
						logrus.Errorf("NotificationScheduler: failed to send email reminder to %s: %v", to, err)
					} else {
						logrus.Infof("NotificationScheduler: email reminder sent successfully to %s", to)
					}
				}(candidateEmail, candidateName, interviewTitle, scheduledTimeStr, meetingLink)
			}

			// Mark as sent
			if err := s.repo.MarkInterview24hReminderSent(ctx, iv.ID); err != nil {
				logrus.Errorf("NotificationScheduler: failed to mark candidate 24h reminder sent for interview %s: %v", iv.ID, err)
			}
		}
	}

	// --- Admin: 1 hour before ---
	adminTo := now.Add(1 * time.Hour)
	adminInterviews, err := s.repo.GetUpcomingInterviewsFor1hReminder(ctx, now, adminTo)
	if err != nil {
		logrus.Errorf("NotificationScheduler: error fetching admin interview reminders: %v", err)
	} else {
		logrus.Infof("NotificationScheduler: admin checks finished, found %d upcoming interviews", len(adminInterviews))
		for _, iv := range adminInterviews {
			if iv.AdminUserID == nil {
				continue
			}
			stageName := ""
			if iv.Stage != nil {
				stageName = iv.Stage.Name
			}
			title := "Interview Reminder"
			desc := fmt.Sprintf("Interview in 1 hour. Stage: %s", stageName)
			if iv.Title != "" {
				desc = fmt.Sprintf("Interview in 1 hour: %s (Stage: %s)", iv.Title, stageName)
			}
			helpers.SendNotificationAsync(ctx, s.repo, *iv.AdminUserID, title, desc)

			// Mark as sent
			if err := s.repo.MarkInterview1hReminderSent(ctx, iv.ID); err != nil {
				logrus.Errorf("NotificationScheduler: failed to mark admin 1h reminder sent for interview %s: %v", iv.ID, err)
			}
		}
	}
}

// checkExpiringContracts sends a review reminder to the employee when a candidate contract expires today.
func (s *NotificationScheduler) checkExpiringContracts(ctx context.Context) {
	logrus.Info("NotificationScheduler: checking expiring contracts...")
	today := time.Now()
	histories, err := s.repo.GetExpiringContractsForNotification(ctx, today)
	if err != nil {
		logrus.Errorf("NotificationScheduler: error fetching expiring contracts: %v", err)
		return
	}
	logrus.Infof("NotificationScheduler: expiring contract checks finished, found %d expiring contracts", len(histories))

	for _, oh := range histories {
		if oh.CandidateUser == nil || oh.Snapshot == nil {
			continue
		}

		candidateName := oh.CandidateUser.Name
		employeeID := extractEmployeeIDFromSnapshot(*oh.Snapshot)
		if employeeID == "" {
			continue
		}

		title := "Contract Expiry - Review Required"
		desc := fmt.Sprintf(
			"The contract of %s in your team has ended. Please fill in the performance review on the candidate's profile page.",
			candidateName,
		)
		helpers.SendNotificationAsync(ctx, s.repo, employeeID, title, desc)
	}
}

// extractEmployeeIDFromSnapshot parses the snapshot JSON string to retrieve employee_user_id.
// The snapshot format: {"employee_user_id":"uuid","employee_name":"...","project_name":"..."}
func extractEmployeeIDFromSnapshot(snapshot string) string {
	const key = `"employee_user_id":"`
	idx := strings.Index(snapshot, key)
	if idx < 0 {
		return ""
	}
	start := idx + len(key)
	rest := snapshot[start:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return ""
	}
	return rest[:end]
}
