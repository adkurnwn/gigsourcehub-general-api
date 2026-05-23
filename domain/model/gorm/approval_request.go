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

// --- Filter ---

type ApprovalRequestFilter struct {
	DefaultFilter
	TableNameEq *string
	StatusEq    *string
	RecordID    *string
}

func (f *ApprovalRequestFilter) Query(q *gorm.DB) {
	f.DefaultFilter.DefaultQuery(q)

	if f.TableNameEq != nil {
		q.Where("table_name = ?", *f.TableNameEq)
	}
	if f.StatusEq != nil {
		q.Where("status = ?", *f.StatusEq)
	}
	if f.RecordID != nil {
		q.Where("record_id = ?", *f.RecordID)
	}
}

// --- Response ---

type ApprovalRequestResp struct {
	ID                     string    `json:"id"`
	RequestedByAdminID     string    `json:"requested_by_admin_id"`
	RequestedByAdminName   *string   `json:"requested_by_admin_name,omitempty"`
	TableName              string    `json:"table_name"`
	RecordID               string    `json:"record_id"`
	Action                 string    `json:"action"`
	ProposedData           *string   `json:"proposed_data"`
	Status                 string    `json:"status"`
	RejectedReason         *string   `json:"rejected_reason"`
	ReviewedBySuperadminID *string   `json:"reviewed_by_superadmin_id"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (row *ApprovalRequest) ToApprovalRequestResp() ApprovalRequestResp {
	var adminName *string
	if row.RequestedByAdmin != nil {
		adminName = &row.RequestedByAdmin.Name
	}

	return ApprovalRequestResp{
		ID:                     row.ID,
		RequestedByAdminID:     row.RequestedByAdminID,
		RequestedByAdminName:   adminName,
		TableName:              row.TableName,
		RecordID:               row.RecordID,
		Action:                 row.Action,
		ProposedData:           row.ProposedData,
		Status:                 row.Status,
		RejectedReason:         row.RejectedReason,
		ReviewedBySuperadminID: row.ReviewedBySuperadminID,
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
	}
}
