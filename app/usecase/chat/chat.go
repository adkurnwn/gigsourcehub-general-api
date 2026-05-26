package usecase_chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	storage_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// excludedStatuses lists recruitment statuses that block conversation creation.
var excludedStatuses = map[string]bool{
	"Unavailable": true,
	"Reject":      true,
	"Cancelled":   true,
}

func (u *appUsecase) CreateConversation(ctx context.Context, adminID string, req request_model.CreateConversationRequest) response.Base {
	return u.createConversation(ctx, adminID, req, false)
}

func (u *appUsecase) StartConversation(ctx context.Context, adminID string, req request_model.CreateConversationRequest) response.Base {
	return u.createConversation(ctx, adminID, req, true)
}

func (u *appUsecase) UploadOffering(ctx context.Context, adminID string, conversationID string, fileHeader *multipart.FileHeader) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if filepath.Ext(strings.ToLower(fileHeader.Filename)) != ".pdf" {
		return response.Error(http.StatusBadRequest, "invalid file type: only PDF files are allowed")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return response.Error(http.StatusBadRequest, "failed to open file")
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return response.Error(http.StatusBadRequest, "failed to read file")
	}

	if detected := http.DetectContentType(fileBytes); detected != "application/pdf" {
		return response.Error(http.StatusBadRequest, "invalid file content: only PDF files are allowed")
	}

	if u.storageRepo == nil {
		return response.Error(http.StatusInternalServerError, "storage service is not available")
	}

	conv, err := u.gormDbRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return response.Error(http.StatusNotFound, "conversation not found")
	}
	if conv.AdminUserID != adminID && conv.CandidateUserID != adminID {
		return response.Error(http.StatusForbidden, "You are not a participant of this conversation")
	}

	objectKey := fmt.Sprintf("offering/%s/%s.pdf", adminID, uuid.NewString())
	expires := 24 * time.Hour
	uploadData, err := u.storageRepo.UploadFilePrivate(objectKey, bytes.NewReader(fileBytes), "application/pdf", &expires)
	if err != nil {
		logrus.Error("UploadOffering S3 error: ", err)
		return response.Error(http.StatusInternalServerError, "failed to upload offering file")
	}

	if uploadData == nil {
		uploadData = &storage_model.UploadResponse{
			Key:         objectKey,
			ContentType: "application/pdf",
			URL:         u.storageRepo.GetPresignedLink(objectKey, &expires),
		}
	}
	uploadData.Filename = fileHeader.Filename
	uploadData.FileSize = fileHeader.Size

	contentBytes, err := json.Marshal(uploadData)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "failed to serialize offering payload")
	}

	msg := gorm_model.Message{
		ConversationID: conversationID,
		Content:        string(contentBytes),
		SenderUserID:   adminID,
	}
	if err := u.gormDbRepo.CreateMessage(ctx, &msg); err != nil {
		logrus.Error("UploadOffering CreateMessage error: ", err)
		return response.Error(http.StatusInternalServerError, "failed to create offering message")
	}

	msgLoaded, err := u.gormDbRepo.GetMessageByID(ctx, msg.ID)
	if err == nil {
		msg = *msgLoaded
	}

	if u.hub != nil {
		var targetUserID string
		var targetRole string
		if conv.AdminUserID == adminID {
			targetUserID = conv.CandidateUserID
			targetRole = "Candidate"
		} else {
			targetUserID = conv.AdminUserID
			targetRole = "Admin"
		}

		broadcastResp := msg.ToMessageResp(targetRole)
		u.hub.SendToUser(targetUserID, "new_message", broadcastResp)
	}

	return response.Success(msg.ToMessageResp("Admin"))
}

func (u *appUsecase) createConversation(ctx context.Context, adminID string, req request_model.CreateConversationRequest, markContacted bool) response.Base {
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
	

	// 5. Check if candidate already has a conversation
	activeConv, err := u.gormDbRepo.GetActiveConversationByCandidateID(ctx, req.CandidateUserID)
	if err == nil && activeConv != nil {
		// If it's for the same subrequest, return it (idempotency)
		if activeConv.SubrequestID == req.SubrequestID {
			if markContacted {
				if err := u.markCandidateContacted(ctx, req.CandidateUserID); err != nil {
					return response.Error(http.StatusInternalServerError, "Failed to update candidate status")
				}
			}
			return response.Success(activeConv.ToConversationResp(userRole))
		}
		// If it's for a different subrequest, they are not allowed to start a new one
		return response.Error(http.StatusBadRequest, "Candidate already has a conversation")
	}

	// 6. Create new conversation
	conv := gorm_model.Conversation{
		AdminUserID:     adminID,
		CandidateUserID: req.CandidateUserID,
		SubrequestID:    req.SubrequestID,
	}
	if markContacted {
		var contactedStatus gorm_model.RecruitmentStatus
		if err := u.gormDbRepo.GetDB().WithContext(ctx).Where("name = ?", "Contacted").First(&contactedStatus).Error; err != nil {
			return response.Error(http.StatusInternalServerError, "Contacted recruitment status not found")
		}

		if err := u.gormDbRepo.StartConversation(ctx, &conv, contactedStatus.ID); err != nil {
			logrus.Error("StartConversation error: ", err)
			return response.Error(http.StatusInternalServerError, "Failed to start conversation")
		}
	} else if err := u.gormDbRepo.CreateConversation(ctx, &conv); err != nil {
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

func (u *appUsecase) markCandidateContacted(ctx context.Context, candidateID string) error {
	statusName, err := u.gormDbRepo.GetCandidateRecruitmentStatusName(ctx, candidateID)
	if err != nil {
		return err
	}

	if statusName == "Contacted" {
		return nil
	}
	if statusName != "Assigned" {
		return nil
	}

	var contactedStatus gorm_model.RecruitmentStatus
	if err := u.gormDbRepo.GetDB().WithContext(ctx).Where("name = ?", "Contacted").First(&contactedStatus).Error; err != nil {
		return err
	}

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: candidateID},
	})
	if err != nil {
		return err
	}
	if user == nil {
		return gorm.ErrRecordNotFound
	}

	user.RecruitmentStatusId = &contactedStatus.ID
	return u.gormDbRepo.UpdateUser(ctx, user)
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
		resp := conv.ToConversationResp(userRole)
		unread, _ := u.gormDbRepo.CountUnreadMessagesByConversation(ctx, conv.ID, userID)
		resp.UnreadCount = unread
		results = append(results, resp)
	}

	var nextCursor *string
	if offset+limit < total {
		nextStr := strconv.FormatInt(page+1, 10)
		nextCursor = &nextStr
	}

	unreadTotal, _ := u.gormDbRepo.CountUnreadMessagesByUser(ctx, userID)

	return response.Success(map[string]interface{}{
		"list":         results,
		"limit":        limit,
		"page":         page,
		"total":        total,
		"cursor":       nextCursor,
		"unread_total": unreadTotal,
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
