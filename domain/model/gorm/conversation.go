package gorm_model

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"time"

	"gorm.io/gorm"
)

type Conversation struct {
	ID              string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	AdminUserID     string         `gorm:"column:admin_user_id;type:uuid;not null"`
	AdminUser       *User          `gorm:"foreignKey:AdminUserID"`
	CandidateUserID string         `gorm:"column:candidate_user_id;type:uuid;not null"`
	CandidateUser   *User          `gorm:"foreignKey:CandidateUserID"`
	SubrequestID    string         `gorm:"column:subrequest_id;type:uuid;not null"`
	Subrequest      *Subrequest    `gorm:"foreignKey:SubrequestID"`
	Messages        []Message      `gorm:"foreignKey:ConversationID"`
	LastMessage     *Message       `gorm:"-"` // populated manually, not a DB relation
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *Conversation) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

type ConversationResp struct {
	ID                          string       `json:"id"`
	SubrequestID                string       `json:"subrequest_id,omitempty"`
	AdminUserID                 string       `json:"admin_user_id,omitempty"`
	AdminUserName               string       `json:"admin_user_name,omitempty"`
	AdminUserProfilePicture     *string      `json:"admin_user_profile_picture,omitempty"`
	CandidateUserID             string       `json:"candidate_user_id"`
	CandidateUserName           string       `json:"candidate_user_name"`
	CandidateUserProfilePicture *string      `json:"candidate_user_profile_picture"`
	LastMessage                 *MessageResp `json:"last_message"`
	UnreadCount                 int64        `json:"unread_count"`
	CreatedAt                   time.Time    `json:"created_at"`
	UpdatedAt                   time.Time    `json:"updated_at"`
}

func (row *Conversation) ToConversationResp(requesterRole string) ConversationResp {
	resp := ConversationResp{
		ID:              row.ID,
		CandidateUserID: row.CandidateUserID,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}

	if requesterRole != "Candidate" {
		resp.SubrequestID = row.SubrequestID
		resp.AdminUserID = row.AdminUserID
		if row.AdminUser != nil {
			resp.AdminUserName = row.AdminUser.Name
			resp.AdminUserProfilePicture = buildProfilePictureURL(row.AdminUser.ProfilePicture)
		}
	}

	if row.CandidateUser != nil {
		resp.CandidateUserName = row.CandidateUser.Name
		resp.CandidateUserProfilePicture = buildProfilePictureURL(row.CandidateUser.ProfilePicture)
	}
	if row.LastMessage != nil {
		msgResp := row.LastMessage.ToMessageResp(requesterRole)
		resp.LastMessage = &msgResp
	}

	return resp
}

func buildProfilePictureURL(pp *string) *string {
	if pp == nil || *pp == "" {
		return nil
	}
	v := *pp
	if len(v) > 0 && v[0] != 'h' {
		v = fmt.Sprintf("%s/%s", os.Getenv("S3_PUBLIC_URL"), v)
	}
	return &v
}
