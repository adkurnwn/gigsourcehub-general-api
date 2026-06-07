package helpers

import (
	"context"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

type WSBroadcaster interface {
	SendToUser(userID string, event string, payload interface{})
}

var wsBroadcaster WSBroadcaster

func SetWSBroadcaster(b WSBroadcaster) {
	wsBroadcaster = b
}

// SendNotification creates a single in-app notification for a specific user.
// It runs synchronously and returns an error if the insert fails.
func SendNotification(ctx context.Context, repo domain.GormRepo, userID, title, description string) error {
	desc := description
	notif := &gorm_model.Notification{
		AdminUserIDOwner: userID,
		Title:            title,
		Description:      &desc,
		IsRead:           false,
		IsAdminBroadcast: false,
	}
	err := repo.CreateNotification(ctx, notif)
	if err == nil && wsBroadcaster != nil {
		wsBroadcaster.SendToUser(userID, "new_notification", notif.ToNotificationResp())
	}
	return err
}

// SendNotificationAsync creates a single notification in a goroutine (fire-and-forget).
func SendNotificationAsync(ctx context.Context, repo domain.GormRepo, userID, title, description string) {
	go func() {
		bgCtx := context.Background()
		if err := SendNotification(bgCtx, repo, userID, title, description); err != nil {
			logrus.Errorf("SendNotificationAsync failed for user %s: %v", userID, err)
		}
	}()
}

// SendNotificationToAll sends a notification to multiple users (fire-and-forget goroutine per user).
func SendNotificationToAll(ctx context.Context, repo domain.GormRepo, userIDs []string, title, description string) {
	for _, uid := range userIDs {
		userID := uid // capture loop variable
		go func() {
			bgCtx := context.Background()
			if err := SendNotification(bgCtx, repo, userID, title, description); err != nil {
				logrus.Errorf("SendNotificationToAll failed for user %s: %v", userID, err)
			}
		}()
	}
}
