package gormrepo

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) FetchFAQ(ctx context.Context, options gorm_model.FAQFilter) ([]gorm_model.FAQ, error) {
	var faqs []gorm_model.FAQ
	q := r.db.WithContext(ctx).Model(&gorm_model.FAQ{})
	options.Query(q)

	if err := q.Find(&faqs).Error; err != nil {
		logrus.Error("FetchFAQ error:", err)
		return nil, err
	}
	return faqs, nil
}

func (r *gormRepo) CountFAQ(ctx context.Context, options gorm_model.FAQFilter) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&gorm_model.FAQ{})

	// Count without limit/offset
	countFilter := options
	countFilter.Limit = nil
	countFilter.Offset = nil
	countFilter.Query(q)

	if err := q.Count(&count).Error; err != nil {
		logrus.Error("CountFAQ error:", err)
		return 0, err
	}
	return count, nil
}

func (r *gormRepo) GetFAQByID(ctx context.Context, id string) (*gorm_model.FAQ, error) {
	var faq gorm_model.FAQ
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&faq).Error
	if err != nil {
		return nil, err
	}
	return &faq, nil
}

func (r *gormRepo) CreateFAQ(ctx context.Context, model *gorm_model.FAQ) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) UpdateFAQ(ctx context.Context, model *gorm_model.FAQ) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) DeleteFAQ(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.FAQ{}, "id = ?", id).Error
}
