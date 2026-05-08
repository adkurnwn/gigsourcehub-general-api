package gormrepo

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
)

func (r *gormRepo) CreateAdminNote(ctx context.Context, model *gorm_model.AdminNote) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) FetchAdminNotesByCandidate(ctx context.Context, candidateID string) ([]gorm_model.AdminNote, error) {
	var notes []gorm_model.AdminNote
	err := r.db.WithContext(ctx).
		Preload("AdminUser").
		Where("candidate_user_id = ?", candidateID).
		Order("created_at DESC").
		Find(&notes).Error
	if err != nil {
		return nil, err
	}

	return notes, nil
}
