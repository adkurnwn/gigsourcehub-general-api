package usecase_aichat

import (
	"context"
	"net/http"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
)

func (u *appUsecase) FetchMyChats(ctx context.Context, adminID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	chats, err := u.gormDbRepo.FetchAIChatsByAdmin(ctx, adminID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch chat sessions: "+err.Error())
	}

	return response.Success(chats)
}

func (u *appUsecase) CreateChat(ctx context.Context, adminID string, firstQuery string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	title := firstQuery
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	chat := gorm_model.AIChat{
		AdminUserID: adminID,
		Title:       title,
	}

	if err := u.gormDbRepo.CreateAIChat(ctx, &chat); err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to create chat session: "+err.Error())
	}

	// We can optionally store the first message here if we want, 
	// but the UI usually handles sending the first message immediately after creation.
	return response.Success(chat)
}

func (u *appUsecase) DeleteChat(ctx context.Context, adminID string, chatID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Verify ownership before delete
	existing, err := u.gormDbRepo.GetAIChatByID(ctx, chatID)
	if err != nil {
		return response.Error(http.StatusNotFound, "Chat session not found")
	}

	if existing.AdminUserID != adminID {
		return response.Error(http.StatusForbidden, "You do not have permission to delete this chat session")
	}

	if err := u.gormDbRepo.DeleteAIChat(ctx, chatID); err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to delete chat session: "+err.Error())
	}

	return response.Success(nil)
}

func (u *appUsecase) FetchChatMessages(ctx context.Context, adminID string, chatID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Verify ownership
	chat, err := u.gormDbRepo.GetAIChatByID(ctx, chatID)
	if err != nil {
		return response.Error(http.StatusNotFound, "Chat session not found")
	}

	if chat.AdminUserID != adminID {
		return response.Error(http.StatusForbidden, "You do not have permission to view this chat session")
	}

	msgs, err := u.gormDbRepo.FetchAIMessagesByChat(ctx, chatID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch chat messages: "+err.Error())
	}

	return response.Success(msgs)
}

func (u *appUsecase) StoreChatMessage(ctx context.Context, adminID string, chatID string, role string, content string, isLast bool) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Verify ownership
	chat, err := u.gormDbRepo.GetAIChatByID(ctx, chatID)
	if err != nil {
		return response.Error(http.StatusNotFound, "Chat session not found")
	}

	if chat.AdminUserID != adminID {
		return response.Error(http.StatusForbidden, "You do not have permission to add messages to this chat session")
	}

	msg := gorm_model.AIMessage{
		AIChatID:      chatID,
		Role:          role,
		Content:       content,
		IsLastMessage: isLast,
	}

	if err := u.gormDbRepo.CreateAIMessage(ctx, &msg); err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to store message: "+err.Error())
	}

	return response.Success(msg)
}
