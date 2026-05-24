package gormrepo

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) GetInterviewByID(ctx context.Context, id string) (*gorm_model.Interview, error) {
	var interview gorm_model.Interview
	err := r.db.WithContext(ctx).
		Preload("Stage").
		Preload("Subrequest").
		Preload("Subrequest.Request").
		Preload("Subrequest.JobRole").
		Preload("CandidateUser").
		Where("id = ?", id).
		First(&interview).Error
	if err != nil {
		return nil, err
	}
	return &interview, nil
}

func (r *gormRepo) CreateInterview(ctx context.Context, model *gorm_model.Interview) error {
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		logrus.Error("CreateInterview error:", err)
		return err
	}
	return nil
}

func (r *gormRepo) UpdateInterview(ctx context.Context, model *gorm_model.Interview) error {
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		logrus.Error("UpdateInterview error:", err)
		return err
	}
	return nil
}

func (r *gormRepo) PatchInterviewStage(ctx context.Context, interviewID string, stageID string, status *string) error {
	updates := map[string]interface{}{
		"stage_id": stageID,
	}
	if status != nil {
		updates["status"] = *status
	}

	if err := r.db.WithContext(ctx).
		Model(&gorm_model.Interview{}).
		Where("id = ?", interviewID).
		Updates(updates).Error; err != nil {
		logrus.Error("PatchInterviewStage error:", err)
		return err
	}
	return nil
}

func (r *gormRepo) PatchInterviewStatus(ctx context.Context, interviewID string, status string) error {
	if err := r.db.WithContext(ctx).
		Model(&gorm_model.Interview{}).
		Where("id = ?", interviewID).
		Update("status", status).Error; err != nil {
		logrus.Error("PatchInterviewStatus error:", err)
		return err
	}
	return nil
}

