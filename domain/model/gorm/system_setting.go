package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type SystemSetting struct {
	ID              string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	IsAIModeEnabled bool           `gorm:"column:is_ai_mode_enabled;type:boolean;default:false"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
