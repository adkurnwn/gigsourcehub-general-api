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
	repo domain.GormRepo
}

func NewNotificationScheduler(repo domain.GormRepo) *NotificationScheduler {
	return &NotificationScheduler{repo: repo}
}

// Start launches the scheduler. It blocks until ctx is cancelled.
// Run in a goroutine: go scheduler.Start(ctx)
func (s *NotificationScheduler) Start(ctx context.Context) {
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
//   - 1-day reminder to candidates (interviews scheduled in the next 23–25 hours)
//   - 1-hour reminder to admins (interviews scheduled in the next 45–75 minutes)
func (s *NotificationScheduler) checkInterviewReminders(ctx context.Context) {
	now := time.Now()

	// --- Candidate: 1 day (24h) before ---
	candidateFrom := now.Add(23 * time.Hour)
	candidateTo := now.Add(25 * time.Hour)
	candidateInterviews, err := s.repo.GetUpcomingInterviewsForNotification(ctx, candidateFrom, candidateTo)
	if err != nil {
		logrus.Errorf("NotificationScheduler: error fetching candidate interview reminders: %v", err)
	} else {
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
		}
	}

	// --- Admin: 1 hour before ---
	adminFrom := now.Add(45 * time.Minute)
	adminTo := now.Add(75 * time.Minute)
	adminInterviews, err := s.repo.GetUpcomingInterviewsForNotification(ctx, adminFrom, adminTo)
	if err != nil {
		logrus.Errorf("NotificationScheduler: error fetching admin interview reminders: %v", err)
	} else {
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
		}
	}
}

// checkExpiringContracts sends a review reminder to the employee when a candidate contract expires today.
func (s *NotificationScheduler) checkExpiringContracts(ctx context.Context) {
	today := time.Now()
	histories, err := s.repo.GetExpiringContractsForNotification(ctx, today)
	if err != nil {
		logrus.Errorf("NotificationScheduler: error fetching expiring contracts: %v", err)
		return
	}

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
