package gormrepo

import (
	"context"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchOnboardHistoriesByCandidate(ctx context.Context, candidateID string, limit, offset int64) ([]gorm_model.OnboardHistory, error) {
	var rows []gorm_model.OnboardHistory

	err := r.db.WithContext(ctx).
		Model(&gorm_model.OnboardHistory{}).
		Preload("CandidateUser").
		Preload("CandidateUser.SystemRole").
		Where("candidate_user_id = ?", candidateID).
		Order("start_date DESC NULLS LAST, created_at DESC").
		Limit(int(limit)).Offset(int(offset)).
		Find(&rows).Error
	if err != nil {
		logrus.Errorf("FetchOnboardHistoriesByCandidate DB Error: %v", err)
		return nil, err
	}

	return rows, nil
}

func (r *gormRepo) CountOnboardHistoriesByCandidate(ctx context.Context, candidateID string) (int64, error) {
	var total int64

	err := r.db.WithContext(ctx).
		Model(&gorm_model.OnboardHistory{}).
		Where("candidate_user_id = ?", candidateID).
		Count(&total).Error
	if err != nil {
		logrus.Errorf("CountOnboardHistoriesByCandidate DB Error: %v", err)
		return 0, err
	}

	return total, nil
}

func (r *gormRepo) FetchOnboardHistoriesByEmployee(ctx context.Context, employeeID string, limit, offset int64) ([]gorm_model.OnboardHistory, error) {
	var rows []gorm_model.OnboardHistory

	err := r.db.WithContext(ctx).
		Model(&gorm_model.OnboardHistory{}).
		Preload("CandidateUser").
		Preload("CandidateUser.SystemRole").
		Where("snapshot ->> 'employee_user_id' = ?", employeeID).
		Order("start_date DESC NULLS LAST, created_at DESC").
		Limit(int(limit)).Offset(int(offset)).
		Find(&rows).Error
	if err != nil {
		logrus.Errorf("FetchOnboardHistoriesByEmployee DB Error: %v", err)
		return nil, err
	}

	return rows, nil
}

func (r *gormRepo) CountOnboardHistoriesByEmployee(ctx context.Context, employeeID string) (int64, error) {
	var total int64

	err := r.db.WithContext(ctx).
		Model(&gorm_model.OnboardHistory{}).
		Where("snapshot ->> 'employee_user_id' = ?", employeeID).
		Count(&total).Error
	if err != nil {
		logrus.Errorf("CountOnboardHistoriesByEmployee DB Error: %v", err)
		return 0, err
	}

	return total, nil
}

func (r *gormRepo) FetchOnboardHistoriesByEmployeeHistory(ctx context.Context, employeeID string, beforeDate time.Time, limit, offset int64) ([]gorm_model.OnboardHistory, error) {
	var rows []gorm_model.OnboardHistory

	err := r.db.WithContext(ctx).
		Model(&gorm_model.OnboardHistory{}).
		Preload("CandidateUser").
		Preload("CandidateUser.SystemRole").
		Where("snapshot ->> 'employee_user_id' = ?", employeeID).
		Where("end_date < ?", beforeDate).
		Order("start_date DESC NULLS LAST, created_at DESC").
		Limit(int(limit)).Offset(int(offset)).
		Find(&rows).Error
	if err != nil {
		logrus.Errorf("FetchOnboardHistoriesByEmployeeHistory DB Error: %v", err)
		return nil, err
	}

	return rows, nil
}

func (r *gormRepo) CountOnboardHistoriesByEmployeeHistory(ctx context.Context, employeeID string, beforeDate time.Time) (int64, error) {
	var total int64

	err := r.db.WithContext(ctx).
		Model(&gorm_model.OnboardHistory{}).
		Where("snapshot ->> 'employee_user_id' = ?", employeeID).
		Where("end_date < ?", beforeDate).
		Count(&total).Error
	if err != nil {
		logrus.Errorf("CountOnboardHistoriesByEmployeeHistory DB Error: %v", err)
		return 0, err
	}

	return total, nil
}
