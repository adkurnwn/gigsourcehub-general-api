package gorm_model

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID             string              `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	ConversationID string              `gorm:"column:conversation_id;type:uuid;not null"`
	Conversation   *Conversation       `gorm:"foreignKey:ConversationID"`
	Content        string              `gorm:"column:content;type:text;not null"`
	SenderUserID       string              `gorm:"column:sender_user_id;type:uuid;not null"`
	SenderUser         *User               `gorm:"foreignKey:SenderUserID"`
	ReplyToMessageID   *string             `gorm:"column:reply_to_message_id;type:uuid"`
	ReplyToMessage     *Message            `gorm:"foreignKey:ReplyToMessageID"`
	Attachments        []MessageAttachment `gorm:"foreignKey:MessageID"`
	ReadAt             *time.Time          `gorm:"column:read_at;type:timestamptz"`
	CreatedAt          time.Time           `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time           `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt          gorm.DeletedAt      `gorm:"column:deleted_at;index"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

type MessageResp struct {
	ID                   string       `json:"id"`
	ConversationID       string       `json:"conversation_id"`
	Content              string       `json:"content"`
	SenderUserID         string       `json:"sender_user_id"`
	SenderName           string       `json:"sender_name,omitempty"`
	SenderProfilePicture *string      `json:"sender_profile_picture,omitempty"`
	ReplyToMessageID     *string      `json:"reply_to_message_id,omitempty"`
	ReplyTo              *ReplyToResp `json:"reply_to,omitempty"`
	ReadAt               *time.Time   `json:"read_at"`
	CreatedAt            time.Time    `json:"created_at"`
}

type ReplyToResp struct {
	ID         string `json:"id"`
	Content    string `json:"content"`
	SenderName string `json:"sender_name"`
}

func (row *Message) ToMessageResp(requesterRole string) MessageResp {
	resp := MessageResp{
		ID:               row.ID,
		ConversationID:   row.ConversationID,
		Content:          row.Content,
		SenderUserID:     row.SenderUserID,
		ReplyToMessageID: row.ReplyToMessageID,
		ReadAt:           row.ReadAt,
		CreatedAt:        row.CreatedAt,
	}

	if row.ReplyToMessage != nil {
		senderName := "Unknown"
		if row.ReplyToMessage.SenderUser != nil {
			senderName = row.ReplyToMessage.SenderUser.Name
			// For candidates, hide non-candidate sender names if requested (optional logic)
		}
		resp.ReplyTo = &ReplyToResp{
			ID:         row.ReplyToMessage.ID,
			Content:    row.ReplyToMessage.Content,
			SenderName: senderName,
		}
	}

	if row.SenderUser != nil {
		isSenderCandidate := false
		if row.SenderUser.SystemRole != nil && row.SenderUser.SystemRole.Name == "Candidate" {
			isSenderCandidate = true
		}

		if requesterRole != "Candidate" || isSenderCandidate {
			resp.SenderName = row.SenderUser.Name
			if row.SenderUser.ProfilePicture != nil && *row.SenderUser.ProfilePicture != "" {
				pp := *row.SenderUser.ProfilePicture
				if len(pp) > 0 && pp[0] != 'h' {
					pp = fmt.Sprintf("%s/%s", os.Getenv("S3_PUBLIC_URL"), pp)
				}
				resp.SenderProfilePicture = &pp
			}
		}
	}

	return resp
}

