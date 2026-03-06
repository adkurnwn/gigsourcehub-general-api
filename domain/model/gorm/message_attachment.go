package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type MessageAttachment struct {
	ID             string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	MessageID      string         `gorm:"column:message_id;type:uuid;not null"`
	Message        *Message       `gorm:"foreignKey:MessageID"`
	AttachmentType string         `gorm:"column:attachment_type;type:message_attachment_type;not null"`
	InterviewID    *string        `gorm:"column:interview_id;type:uuid"`
	Interview      *Interview     `gorm:"foreignKey:InterviewID"`
	ContractID     *string        `gorm:"column:contract_id;type:uuid"`
	Contract       *Contract      `gorm:"foreignKey:ContractID"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
