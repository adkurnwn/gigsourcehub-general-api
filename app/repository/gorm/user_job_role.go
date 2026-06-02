package gormrepo

import (
	"context"
	"fmt"
)

// GetJobRolesByUserIDs fetches names of job roles for a list of user IDs from user_has_job_roles table.
func (r *gormRepo) GetJobRolesByUserIDs(ctx context.Context, userIDs []string) (map[string][]string, error) {
	if len(userIDs) == 0 {
		return map[string][]string{}, nil
	}

	type row struct {
		UserID       string `gorm:"column:user_id"`
		JobRoleName  string `gorm:"column:job_role_name"`
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("user_has_job_roles uhjr").
		Select("uhjr.user_id, jr.name as job_role_name").
		Joins("JOIN job_roles jr ON uhjr.job_role_id = jr.id").
		Where("uhjr.user_id IN (?)", userIDs).
		Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("GetJobRolesByUserIDs query error: %w", err)
	}

	result := make(map[string][]string)
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], row.JobRoleName)
	}

	return result, nil
}
