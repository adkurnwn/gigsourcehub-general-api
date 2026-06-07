package gormrepo

import (
	"context"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
)

// CreateNotification persists a new notification record.
func (r *gormRepo) CreateNotification(ctx context.Context, model *gorm_model.Notification) error {
	return r.db.WithContext(ctx).Create(model).Error
}

// FetchNotificationsByUser returns paginated notifications for a given user, newest first.
func (r *gormRepo) FetchNotificationsByUser(ctx context.Context, userID string, limit, offset int64) ([]gorm_model.Notification, error) {
	var notifications []gorm_model.Notification
	err := r.db.WithContext(ctx).
		Where("admin_user_id_owner = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Limit(int(limit)).Offset(int(offset)).
		Find(&notifications).Error
	return notifications, err
}

// CountNotificationsByUser counts all notifications for a given user.
func (r *gormRepo) CountNotificationsByUser(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&gorm_model.Notification{}).
		Where("admin_user_id_owner = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// CountUnreadNotificationsByUser counts unread notifications for a given user.
func (r *gormRepo) CountUnreadNotificationsByUser(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&gorm_model.Notification{}).
		Where("admin_user_id_owner = ? AND is_read = false AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// MarkNotificationAsRead marks a single notification as read (only if it belongs to the user).
func (r *gormRepo) MarkNotificationAsRead(ctx context.Context, notificationID, userID string) error {
	return r.db.WithContext(ctx).
		Model(&gorm_model.Notification{}).
		Where("id = ? AND admin_user_id_owner = ?", notificationID, userID).
		Update("is_read", true).Error
}

// MarkAllNotificationsAsRead marks all notifications for a user as read.
func (r *gormRepo) MarkAllNotificationsAsRead(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&gorm_model.Notification{}).
		Where("admin_user_id_owner = ? AND is_read = false AND deleted_at IS NULL", userID).
		Update("is_read", true).Error
}

// GetSuperadminUserIDs returns the IDs of all users with the Superadmin system role.
func (r *gormRepo) GetSuperadminUserIDs(ctx context.Context) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).
		Table("users").
		Joins("JOIN system_roles sr ON sr.id = users.system_role_id").
		Where("sr.name = ? AND users.deleted_at IS NULL", "Superadmin").
		Pluck("users.id", &ids).Error
	return ids, err
}

// GetUpcomingInterviewsForNotification fetches SCHEDULED interviews whose scheduled_at falls between from and to.
func (r *gormRepo) GetUpcomingInterviewsForNotification(ctx context.Context, from, to time.Time) ([]gorm_model.Interview, error) {
	var interviews []gorm_model.Interview
	err := r.db.WithContext(ctx).
		Preload("Stage").
		Where("status = ? AND scheduled_at > ? AND scheduled_at <= ? AND deleted_at IS NULL", "SCHEDULED", from, to).
		Find(&interviews).Error
	return interviews, err
}

// GetExpiringContractsForNotification fetches onboard_histories whose end_date equals targetDate and is_stopped = false.
func (r *gormRepo) GetExpiringContractsForNotification(ctx context.Context, targetDate time.Time) ([]gorm_model.OnboardHistory, error) {
	var histories []gorm_model.OnboardHistory
	dateStr := targetDate.Format("2006-01-02")
	err := r.db.WithContext(ctx).
		Preload("CandidateUser").
		Where("end_date = ? AND is_stopped = false AND deleted_at IS NULL", dateStr).
		Find(&histories).Error
	return histories, err
}

// GetUpcomingInterviewsFor24hReminder fetches SCHEDULED interviews whose scheduled_at falls between from and to, and is_24h_reminder_sent is false.
func (r *gormRepo) GetUpcomingInterviewsFor24hReminder(ctx context.Context, from, to time.Time) ([]gorm_model.Interview, error) {
	var interviews []gorm_model.Interview
	err := r.db.WithContext(ctx).
		Preload("Stage").
		Preload("CandidateUser").
		Where("status = ? AND scheduled_at > ? AND scheduled_at <= ? AND is_24h_reminder_sent = false AND deleted_at IS NULL", "SCHEDULED", from, to).
		Find(&interviews).Error
	return interviews, err
}

// GetUpcomingInterviewsFor1hReminder fetches SCHEDULED interviews whose scheduled_at falls between from and to, and is_1h_reminder_sent is false.
func (r *gormRepo) GetUpcomingInterviewsFor1hReminder(ctx context.Context, from, to time.Time) ([]gorm_model.Interview, error) {
	var interviews []gorm_model.Interview
	err := r.db.WithContext(ctx).
		Preload("Stage").
		Where("status = ? AND scheduled_at > ? AND scheduled_at <= ? AND is_1h_reminder_sent = false AND deleted_at IS NULL", "SCHEDULED", from, to).
		Find(&interviews).Error
	return interviews, err
}

// MarkInterview24hReminderSent sets is_24h_reminder_sent to true for the specified interview.
func (r *gormRepo) MarkInterview24hReminderSent(ctx context.Context, interviewID string) error {
	return r.db.WithContext(ctx).
		Model(&gorm_model.Interview{}).
		Where("id = ?", interviewID).
		Update("is_24h_reminder_sent", true).Error
}

// MarkInterview1hReminderSent sets is_1h_reminder_sent to true for the specified interview.
func (r *gormRepo) MarkInterview1hReminderSent(ctx context.Context, interviewID string) error {
	return r.db.WithContext(ctx).
		Model(&gorm_model.Interview{}).
		Where("id = ?", interviewID).
		Update("is_1h_reminder_sent", true).Error
}
