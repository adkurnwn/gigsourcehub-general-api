package usecase_chat

import (
	"context"
	"net/http"
	"strconv"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/sirupsen/logrus"
)

// excludedStatuses lists recruitment statuses that block conversation creation.
var excludedStatuses = map[string]bool{
	"Unavailable": true,
	"Reject":      true,
	"Cancelled":   true,
}

func (u *appUsecase) CreateConversation(ctx context.Context, adminID string, req request_model.CreateConversationRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	userRole, errRole := u.gormDbRepo.GetRoleNameByUserID(ctx, adminID)
	if errRole != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify requester role")
	}

	// 1. Verify candidate exists and is a Candidate
	candidateUser, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: req.CandidateUserID},
	})
	if err != nil || candidateUser == nil {
		return response.Error(http.StatusNotFound, "Candidate user not found")
	}

	roleName, err := u.gormDbRepo.GetRoleNameByUserID(ctx, req.CandidateUserID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify user role")
	}
	if roleName != "Candidate" {
		return response.Error(http.StatusBadRequest, "Target user is not a Candidate")
	}

	// 2. Verify admin owns the subrequest's parent request
	isAdmin, err := u.gormDbRepo.IsAdminOfSubrequest(ctx, adminID, req.SubrequestID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify admin assignment")
	}
	if !isAdmin {
		return response.Error(http.StatusForbidden, "You are not assigned to this subrequest's request")
	}

	// 3. Verify candidate is assigned to this subrequest
	isCandidate, err := u.gormDbRepo.IsCandidateOnSubrequest(ctx, req.CandidateUserID, req.SubrequestID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify candidate assignment")
	}
	if !isCandidate {
		return response.Error(http.StatusBadRequest, "Candidate is not assigned to this subrequest")
	}

	// 4. Check recruitment status
	statusName, err := u.gormDbRepo.GetCandidateRecruitmentStatusName(ctx, req.CandidateUserID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to check recruitment status")
	}
	if excludedStatuses[statusName] {
		return response.Error(http.StatusBadRequest, "Cannot start conversation: candidate status is "+statusName)
	}

	// 5. Check if conversation already exists for this subrequest + candidate
	existing, _ := u.gormDbRepo.GetConversationBySubrequestAndCandidate(ctx, req.SubrequestID, req.CandidateUserID)
	if existing != nil {
		return response.Success(existing.ToConversationResp(userRole))
	}

	// 6. Create new conversation
	conv := gorm_model.Conversation{
		AdminUserID:     adminID,
		CandidateUserID: req.CandidateUserID,
		SubrequestID:    req.SubrequestID,
	}
	if err := u.gormDbRepo.CreateConversation(ctx, &conv); err != nil {
		logrus.Error("CreateConversation error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to create conversation")
	}

	// Reload with preloads
	created, err := u.gormDbRepo.GetConversationByID(ctx, conv.ID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Conversation created but failed to load")
	}

	return response.Success(created.ToConversationResp(userRole))
}

func (u *appUsecase) FetchMyConversations(ctx context.Context, userID string, page, limit int64) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	userRole, errRole := u.gormDbRepo.GetRoleNameByUserID(ctx, userID)
	if errRole != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify requester role")
	}

	offset := (page - 1) * limit

	total, err := u.gormDbRepo.CountConversationsByUser(ctx, userID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count conversations")
	}

	conversations, err := u.gormDbRepo.FetchConversationsByUser(ctx, userID, limit, offset)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch conversations")
	}

	var results []interface{}
	for _, conv := range conversations {
		results = append(results, conv.ToConversationResp(userRole))
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

func (u *appUsecase) GetConversation(ctx context.Context, userID string, conversationID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	userRole, errRole := u.gormDbRepo.GetRoleNameByUserID(ctx, userID)
	if errRole != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify requester role")
	}

	conv, err := u.gormDbRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return response.Error(http.StatusNotFound, "Conversation not found")
	}

	if conv.AdminUserID != userID && conv.CandidateUserID != userID {
		return response.Error(http.StatusForbidden, "You are not a participant of this conversation")
	}

	return response.Success(conv.ToConversationResp(userRole))
}

func (u *appUsecase) SendMessage(ctx context.Context, userID string, conversationID string, req request_model.SendMessageRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	userRole, errRole := u.gormDbRepo.GetRoleNameByUserID(ctx, userID)
	if errRole != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify requester role")
	}

	// Verify conversation exists and user is a participant
	conv, err := u.gormDbRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return response.Error(http.StatusNotFound, "Conversation not found")
	}
	if conv.AdminUserID != userID && conv.CandidateUserID != userID {
		return response.Error(http.StatusForbidden, "You are not a participant of this conversation")
	}

	msg := gorm_model.Message{
		ConversationID:   conversationID,
		Content:          req.Content,
		SenderUserID:     userID,
		ReplyToMessageID: req.ReplyToMessageID,
	}
	if err := u.gormDbRepo.CreateMessage(ctx, &msg); err != nil {
		logrus.Error("SendMessage error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to send message")
	}

	// Reload message with full info (sender, replies, etc.)
	reloaded, err := u.gormDbRepo.GetMessageByID(ctx, msg.ID)
	if err == nil {
		msg = *reloaded
	}

	msgResp := msg.ToMessageResp(userRole)

	// Broadcast via WebSocket to the other participant
	if u.hub != nil {
		var targetUserID string
		var targetRole string
		if conv.AdminUserID == userID {
			targetUserID = conv.CandidateUserID
			targetRole = "Candidate"
		} else {
			targetUserID = conv.AdminUserID
			targetRole = "Admin" // Or something else, doesn't matter much since it's not candidate
		}
		
		// The message sent to the other user should be formatted for their role!
		broadcastResp := msg.ToMessageResp(targetRole)
		u.hub.SendToUser(targetUserID, "new_message", broadcastResp)
	}

	return response.Success(msgResp)
}

func (u *appUsecase) FetchMessages(ctx context.Context, userID string, conversationID string, page, limit int64) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	userRole, errRole := u.gormDbRepo.GetRoleNameByUserID(ctx, userID)
	if errRole != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify requester role")
	}

	// Verify conversation exists and user is a participant
	conv, err := u.gormDbRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return response.Error(http.StatusNotFound, "Conversation not found")
	}
	if conv.AdminUserID != userID && conv.CandidateUserID != userID {
		return response.Error(http.StatusForbidden, "You are not a participant of this conversation")
	}

	offset := (page - 1) * limit

	total, err := u.gormDbRepo.CountMessagesByConversation(ctx, conversationID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count messages")
	}

	messages, err := u.gormDbRepo.FetchMessagesByConversation(ctx, conversationID, limit, offset)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch messages")
	}

	var results []interface{}
	for _, msg := range messages {
		results = append(results, msg.ToMessageResp(userRole))
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

func (u *appUsecase) MarkAsRead(ctx context.Context, userID string, conversationID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Verify conversation exists and user is a participant
	conv, err := u.gormDbRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return response.Error(http.StatusNotFound, "Conversation not found")
	}
	if conv.AdminUserID != userID && conv.CandidateUserID != userID {
		return response.Error(http.StatusForbidden, "You are not a participant of this conversation")
	}

	if err := u.gormDbRepo.MarkMessagesAsRead(ctx, conversationID, userID); err != nil {
		logrus.Error("MarkAsRead error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to mark messages as read")
	}

	// Broadcast read receipt via WebSocket
	if u.hub != nil {
		var targetUserID string
		if conv.AdminUserID == userID {
			targetUserID = conv.CandidateUserID
		} else {
			targetUserID = conv.AdminUserID
		}
		u.hub.SendToUser(targetUserID, "message_read", map[string]string{
			"conversation_id": conversationID,
			"reader_user_id":  userID,
		})
	}

	return response.Success(nil)
}
