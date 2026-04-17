package gormrepo

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"gorm.io/gorm"
)

func (r *gormRepo) CreateAIChat(ctx context.Context, chat *gorm_model.AIChat) error {
	return r.db.WithContext(ctx).Create(chat).Error
}

func (r *gormRepo) GetAIChatByID(ctx context.Context, id string) (*gorm_model.AIChat, error) {
	var chat gorm_model.AIChat
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&chat).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *gormRepo) FetchAIChatsByAdmin(ctx context.Context, adminID string) ([]gorm_model.AIChat, error) {
	var chats []gorm_model.AIChat
	err := r.db.WithContext(ctx).
		Where("admin_user_id = ?", adminID).
		Order("created_at DESC").
		Find(&chats).Error
	return chats, err
}

func (r *gormRepo) DeleteAIChat(ctx context.Context, id string) error {
	// GORM's Delete on a model with a foreign key like AIChat might not automatically
	// delete associated messages unless configured for cascading, or we do it manually.
	// Since we want to be safe, we'll use a transaction.
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete all messages associated with this chat
		if err := tx.Where("ai_chat_id = ?", id).Delete(&gorm_model.AIMessage{}).Error; err != nil {
			return err
		}
		// Delete the chat session itself
		if err := tx.Where("id = ?", id).Delete(&gorm_model.AIChat{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *gormRepo) CreateAIMessage(ctx context.Context, msg *gorm_model.AIMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *gormRepo) FetchAIMessagesByChat(ctx context.Context, chatID string) ([]gorm_model.AIMessage, error) {
	var msgs []gorm_model.AIMessage
	err := r.db.WithContext(ctx).
		Where("ai_chat_id = ?", chatID).
		Order("created_at ASC").
		Find(&msgs).Error
	return msgs, err
}
