package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogActivity struct {
	ID          string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	ActorID     *string        `gorm:"column:actor_id;type:uuid"`
	Actor       *User          `gorm:"foreignKey:ActorID"`
	ActionType  string         `gorm:"column:action_type;type:varchar(100);not null"`
	Module      string         `gorm:"column:module;type:varchar(100);not null"`
	Description *string        `gorm:"column:description;type:text"`
	Metadata    *string        `gorm:"column:metadata;type:jsonb"`
	IPAddress   *string        `gorm:"column:ip_address;type:varchar(45)"`
	IsSuccess   bool           `gorm:"column:is_success;type:boolean;not null"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *LogActivity) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

type LogActivityFilter struct {
	DefaultFilter
	ActorID    *string
	ActionType *string
	Module     *string
	Search     *string
	StartDate  *time.Time
	EndDate    *time.Time
	RoleName   *string
	IsSuccess  *bool
}

func (f *LogActivityFilter) Query(q *gorm.DB) {
	// default query
	f.DefaultFilter.DefaultQuery(q)

	if f.ActorID != nil {
		q.Where("actor_id = ?", *f.ActorID)
	}
	if f.ActionType != nil {
		q.Where("action_type = ?", *f.ActionType)
	}
	if f.Module != nil {
		q.Where("module = ?", *f.Module)
	}
	if f.StartDate != nil {
		q.Where("created_at >= ?", *f.StartDate)
	}
	if f.EndDate != nil {
		q.Where("created_at <= ?", *f.EndDate)
	}
	if f.RoleName != nil {
		q.Where("EXISTS (SELECT 1 FROM users u JOIN system_roles sr ON u.system_role_id = sr.id WHERE u.id = actor_id AND sr.name = ?)", *f.RoleName)
	}
	if f.IsSuccess != nil {
		q.Where("is_success = ?", *f.IsSuccess)
	}
	if f.Search != nil && *f.Search != "" {
		s := "%" + *f.Search + "%"
		q.Where("(action_type ILIKE ? OR module ILIKE ? OR description ILIKE ? OR ip_address ILIKE ? OR EXISTS (SELECT 1 FROM users u WHERE u.id = actor_id AND u.name ILIKE ?))", s, s, s, s, s)
	}
}
