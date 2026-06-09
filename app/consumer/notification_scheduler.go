package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
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
	// Ticker fires every 15 minutes for contract expiry check
	contractTicker := time.NewTicker(15 * time.Minute)
	// Ticker fires every 15 minutes for candidate availability expiry check
	availabilityTicker := time.NewTicker(15 * time.Minute)

	defer interviewTicker.Stop()
	defer contractTicker.Stop()
	defer availabilityTicker.Stop()

	// Run once immediately on startup so we don't have to wait for the first tick
	s.checkInterviewReminders(ctx)
	s.checkExpiringContracts(ctx)
	s.checkCandidateAvailabilityExpiry(ctx)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("NotificationScheduler: context cancelled, stopping")
			return
		case <-interviewTicker.C:
			s.checkInterviewReminders(ctx)
		case <-contractTicker.C:
			s.checkExpiringContracts(ctx)
		case <-availabilityTicker.C:
			s.checkCandidateAvailabilityExpiry(ctx)
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

// checkExpiringContracts sends a review reminder to the employee when a candidate contract expires today,
// ends the contract by stopping onboarding, and marks the history record.
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
		if oh.CandidateUser == nil {
			logrus.Warnf("NotificationScheduler: skipping onboard history %s because CandidateUser is nil", oh.ID)
			continue
		}
		if oh.Snapshot == nil {
			logrus.Warnf("NotificationScheduler: skipping onboard history %s because Snapshot is nil", oh.ID)
			continue
		}

		candidateName := oh.CandidateUser.Name
		employeeID := extractEmployeeIDFromSnapshot(*oh.Snapshot)
		if employeeID == "" {
			logrus.Warnf("NotificationScheduler: skipping onboard history %s because employee_user_id not found in snapshot: %s", oh.ID, *oh.Snapshot)
			continue
		}

		title := "Contract Expiry - Review Required"
		desc := fmt.Sprintf(
			"The contract of %s in your team has ended. Please fill in the performance review on the candidate's profile page.",
			candidateName,
		)
		helpers.SendNotificationAsync(ctx, s.repo, employeeID, title, desc)

		if oh.CandidateUserID != "" {
			if err := s.repo.EndExpiredContract(ctx, oh.ID, oh.CandidateUserID); err != nil {
				logrus.Errorf("NotificationScheduler: failed to end expired contract for history %s: %v", oh.ID, err)
			} else {
				logrus.Infof("NotificationScheduler: successfully ended contract for history %s, candidate %s", oh.ID, oh.CandidateUserID)
			}
		}
	}
}

// extractEmployeeIDFromSnapshot parses the snapshot JSON string to retrieve employee_user_id.
// The snapshot format: {"employee_user_id":"uuid","employee_name":"...","project_name":"..."}
func extractEmployeeIDFromSnapshot(snapshot string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(snapshot), &data); err == nil {
		if val, ok := data["employee_user_id"].(string); ok {
			return val
		}
	}

	// Fallback to string searching if JSON unmarshaling fails
	const key = `"employee_user_id":"`
	idx := strings.Index(snapshot, key)
	if idx < 0 {
		// Try with space
		const keyWithSpace = `"employee_user_id": "`
		idx = strings.Index(snapshot, keyWithSpace)
		if idx < 0 {
			return ""
		}
		start := idx + len(keyWithSpace)
		rest := snapshot[start:]
		end := strings.Index(rest, `"`)
		if end < 0 {
			return ""
		}
		return rest[:end]
	}
	start := idx + len(key)
	rest := snapshot[start:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return ""
	}
	return rest[:end]
}

// checkCandidateAvailabilityExpiry queries the 'Unavailable' status and updates
// any candidate whose availability date has expired (unavailable_until <= NOW())
// back to 'Available' (NULL recruitment_status_id and NULL unavailable_until).
func (s *NotificationScheduler) checkCandidateAvailabilityExpiry(ctx context.Context) {
	logrus.Info("NotificationScheduler: checking candidate availability expiry...")
	db := s.repo.GetDB()

	var unavailableStatus gorm_model.RecruitmentStatus
	if err := db.WithContext(ctx).
		Where("name = ? AND deleted_at IS NULL", "Unavailable").
		First(&unavailableStatus).Error; err != nil {
		logrus.Errorf("NotificationScheduler: failed to fetch 'Unavailable' status: %v", err)
		return
	}

	now := time.Now()
	res := db.WithContext(ctx).
		Model(&gorm_model.User{}).
		Where("recruitment_status_id = ? AND unavailable_until IS NOT NULL AND unavailable_until <= ?", unavailableStatus.ID, now).
		Updates(map[string]interface{}{
			"recruitment_status_id": nil,
			"unavailable_until":     nil,
		})

	if res.Error != nil {
		logrus.Errorf("NotificationScheduler: failed to update expired candidate availabilities: %v", res.Error)
	} else if res.RowsAffected > 0 {
		logrus.Infof("NotificationScheduler: successfully reset availability for %d candidates", res.RowsAffected)
	}
}
