package gormrepo

import (
	"context"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (r *gormRepo) CreateConversation(ctx context.Context, conv *gorm_model.Conversation) error {
	if err := r.db.WithContext(ctx).Create(conv).Error; err != nil {
		logrus.Errorf("CreateConversation DB Error: %v", err)
		return err
	}
	return nil
}

func (r *gormRepo) GetConversationByID(ctx context.Context, id string) (*gorm_model.Conversation, error) {
	var conv gorm_model.Conversation
	err := r.db.WithContext(ctx).
		Preload("AdminUser").
		Preload("CandidateUser").
		Where("id = ?", id).
		First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *gormRepo) GetConversationBySubrequestAndCandidate(ctx context.Context, subrequestID, candidateID string) (*gorm_model.Conversation, error) {
	var conv gorm_model.Conversation
	err := r.db.WithContext(ctx).
		Preload("AdminUser").
		Preload("CandidateUser").
		Where("subrequest_id = ? AND candidate_user_id = ?", subrequestID, candidateID).
		First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *gormRepo) FetchConversationsByUser(ctx context.Context, userID string, limit, offset int64) ([]gorm_model.Conversation, error) {
	var conversations []gorm_model.Conversation
	err := r.db.WithContext(ctx).
		Preload("AdminUser").
		Preload("CandidateUser").
		Where("admin_user_id = ? OR candidate_user_id = ?", userID, userID).
		Order("updated_at DESC").
		Limit(int(limit)).Offset(int(offset)).
		Find(&conversations).Error
	if err != nil {
		logrus.Errorf("FetchConversationsByUser DB Error: %v", err)
		return nil, err
	}

	// Populate LastMessage for each conversation
	for i := range conversations {
		var lastMsg gorm_model.Message
		err := r.db.WithContext(ctx).
			Preload("SenderUser").
			Where("conversation_id = ?", conversations[i].ID).
			Order("created_at DESC").
			First(&lastMsg).Error
		if err == nil {
			conversations[i].LastMessage = &lastMsg
		}
	}

	return conversations, nil
}

func (r *gormRepo) CountConversationsByUser(ctx context.Context, userID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&gorm_model.Conversation{}).
		Where("admin_user_id = ? OR candidate_user_id = ?", userID, userID).
		Count(&total).Error
	if err != nil {
		logrus.Errorf("CountConversationsByUser DB Error: %v", err)
		return 0, err
	}
	return total, nil
}

func (r *gormRepo) CreateMessage(ctx context.Context, msg *gorm_model.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(msg).Error; err != nil {
			logrus.Errorf("CreateMessage DB Error: %v", err)
			return err
		}

		// Touch conversation updated_at
		if err := tx.Model(&gorm_model.Conversation{}).
			Where("id = ?", msg.ConversationID).
			UpdateColumn("updated_at", time.Now()).Error; err != nil {
			logrus.Errorf("CreateMessage update conversation updated_at Error: %v", err)
			return err
		}

		return nil
	})
}

func (r *gormRepo) GetMessageByID(ctx context.Context, id string) (*gorm_model.Message, error) {
	var msg gorm_model.Message
	err := r.db.WithContext(ctx).
		Preload("SenderUser").
		Preload("ReplyToMessage").
		Preload("ReplyToMessage.SenderUser").
		Where("id = ?", id).
		First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *gormRepo) FetchMessagesByConversation(ctx context.Context, conversationID string, limit, offset int64) ([]gorm_model.Message, error) {
	var messages []gorm_model.Message
	err := r.db.WithContext(ctx).
		Preload("SenderUser").
		Preload("ReplyToMessage").
		Preload("ReplyToMessage.SenderUser").
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(int(limit)).Offset(int(offset)).
		Find(&messages).Error
	if err != nil {
		logrus.Errorf("FetchMessagesByConversation DB Error: %v", err)
		return nil, err
	}
	return messages, nil
}

func (r *gormRepo) CountMessagesByConversation(ctx context.Context, conversationID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&gorm_model.Message{}).
		Where("conversation_id = ?", conversationID).
		Count(&total).Error
	if err != nil {
		logrus.Errorf("CountMessagesByConversation DB Error: %v", err)
		return 0, err
	}
	return total, nil
}

func (r *gormRepo) MarkMessagesAsRead(ctx context.Context, conversationID, readerUserID string) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&gorm_model.Message{}).
		Where("conversation_id = ? AND sender_user_id != ? AND read_at IS NULL", conversationID, readerUserID).
		UpdateColumn("read_at", now).Error
	if err != nil {
		logrus.Errorf("MarkMessagesAsRead DB Error: %v", err)
	}
	return err
}

func (r *gormRepo) CountUnreadMessagesByUser(ctx context.Context, userID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&gorm_model.Message{}).
		Joins("JOIN conversations ON conversations.id = messages.conversation_id").
		Where("(conversations.admin_user_id = ? OR conversations.candidate_user_id = ?) AND messages.sender_user_id != ? AND messages.read_at IS NULL", userID, userID, userID).
		Count(&total).Error
	if err != nil {
		logrus.Errorf("CountUnreadMessagesByUser DB Error: %v", err)
		return 0, err
	}
	return total, nil
}

func (r *gormRepo) CountUnreadMessagesByConversation(ctx context.Context, conversationID, userID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&gorm_model.Message{}).
		Where("conversation_id = ? AND sender_user_id != ? AND read_at IS NULL", conversationID, userID).
		Count(&total).Error
	if err != nil {
		logrus.Errorf("CountUnreadMessagesByConversation DB Error: %v", err)
		return 0, err
	}
	return total, nil
}

func (r *gormRepo) IsAdminOfSubrequest(ctx context.Context, adminID, subrequestID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("subrequests").
		Joins("JOIN requests ON requests.id = subrequests.request_id").
		Where("subrequests.id = ? AND requests.admin_user_id = ? AND subrequests.deleted_at IS NULL AND requests.deleted_at IS NULL", subrequestID, adminID).
		Count(&count).Error
	if err != nil {
		logrus.Errorf("IsAdminOfSubrequest DB Error: %v", err)
		return false, err
	}
	return count > 0, nil
}

func (r *gormRepo) IsCandidateOnSubrequest(ctx context.Context, candidateID, subrequestID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("subrequest_candidates").
		Where("candidate_user_id = ? AND subrequest_id = ? AND deleted_at IS NULL", candidateID, subrequestID).
		Count(&count).Error
	if err != nil {
		logrus.Errorf("IsCandidateOnSubrequest DB Error: %v", err)
		return false, err
	}
	return count > 0, nil
}

func (r *gormRepo) GetCandidateRecruitmentStatusName(ctx context.Context, candidateID string) (string, error) {
	var result struct {
		StatusName string
	}
	err := r.db.WithContext(ctx).
		Table("users").
		Select("recruitment_statuses.name as status_name").
		Joins("LEFT JOIN recruitment_statuses ON recruitment_statuses.id = users.recruitment_status_id").
		Where("users.id = ? AND users.deleted_at IS NULL", candidateID).
		Scan(&result).Error
	if err != nil {
		logrus.Errorf("GetCandidateRecruitmentStatusName DB Error: %v", err)
		return "", err
	}
	return result.StatusName, nil
}
