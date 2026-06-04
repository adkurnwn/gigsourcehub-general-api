package usecase_notification

import (
	"context"
	"net/http"
	"strconv"

	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
)

// FetchMyNotifications returns a paginated list of notifications for the authenticated user.
func (u *appUsecase) FetchMyNotifications(ctx context.Context, userID string, page, limit int64) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit

	total, err := u.gormDbRepo.CountNotificationsByUser(ctx, userID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count notifications")
	}

	notifications, err := u.gormDbRepo.FetchNotificationsByUser(ctx, userID, limit, offset)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch notifications")
	}

	var results []interface{}
	for _, n := range notifications {
		results = append(results, n.ToNotificationResp())
	}

	var nextCursor *string
	if offset+limit < total {
		nextStr := strconv.FormatInt(page+1, 10)
		nextCursor = &nextStr
	}

	return response.Success(response.List{
		List:   results,
		Limit:  limit,
		Page:   page,
		Total:  total,
		Cursor: nextCursor,
	})
}

// MarkAsRead marks a single notification as read for the given user.
func (u *appUsecase) MarkAsRead(ctx context.Context, userID, notificationID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.gormDbRepo.MarkNotificationAsRead(ctx, notificationID, userID); err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to mark notification as read")
	}

	return response.Success(nil)
}

// MarkAllAsRead marks all notifications for the given user as read.
func (u *appUsecase) MarkAllAsRead(ctx context.Context, userID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.gormDbRepo.MarkAllNotificationsAsRead(ctx, userID); err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to mark all notifications as read")
	}

	return response.Success(nil)
}

// GetUnreadCount returns the count of unread notifications for the given user.
func (u *appUsecase) GetUnreadCount(ctx context.Context, userID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	count, err := u.gormDbRepo.CountUnreadNotificationsByUser(ctx, userID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count unread notifications")
	}

	return response.Success(map[string]interface{}{
		"unread_count": count,
	})
}
