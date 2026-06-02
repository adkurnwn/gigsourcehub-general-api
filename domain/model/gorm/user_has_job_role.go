package gorm_model

type UserHasJobRole struct {
	UserID    string `gorm:"column:user_id;primaryKey;type:uuid"`
	JobRoleID string `gorm:"column:job_role_id;primaryKey;type:uuid"`
}
