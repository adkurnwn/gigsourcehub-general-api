package gorm_model

import "time"

type Bookmark struct {
	AdminID     string    `gorm:"column:admin_id;primaryKey;type:uuid"`
	CandidateID string    `gorm:"column:candidate_id;primaryKey;type:uuid"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp"`
}
