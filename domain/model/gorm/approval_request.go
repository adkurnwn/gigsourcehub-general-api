package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApprovalRequest struct {
	ID                     string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	RequestedByAdminID     string         `gorm:"column:requested_by_admin_id;type:uuid;not null"`
	RequestedByAdmin       *User          `gorm:"foreignKey:RequestedByAdminID"`
	TableName              string         `gorm:"column:table_name;type:approval_request_table_name;not null"`
	RecordID               string         `gorm:"column:record_id;type:uuid;not null"`
	Action                 string         `gorm:"column:action;type:approval_request_action;not null"`
	ProposedData           *string        `gorm:"column:proposed_data;type:jsonb"`
	Status                 string         `gorm:"column:status;type:approval_request_status;default:'PENDING';not null"`
	RejectedReason         *string        `gorm:"column:rejected_reason;type:text"`
	ReviewedBySuperadminID *string        `gorm:"column:reviewed_by_superadmin_id;type:uuid"`
	ReviewedBySuperadmin   *User          `gorm:"foreignKey:ReviewedBySuperadminID"`
	CreatedAt              time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt              time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt              gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *ApprovalRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}
