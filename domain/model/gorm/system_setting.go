package gorm_model

import (
	"time"

	"gorm.io/gorm"
)

type SystemSetting struct {
	ID              string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()" json:"id"`
	IsAIModeEnabled bool           `gorm:"column:is_ai_mode_enabled;type:boolean;default:false" json:"is_ai_mode_enabled"`
	CVTemplatePath  *string        `gorm:"column:cv_template_path;type:varchar(255)" json:"cv_template_path,omitempty"`
	CVTemplateURL   string         `gorm:"-" json:"cv_template_url,omitempty"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}
