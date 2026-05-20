package gormrepo

import (
	"context"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
)

func (r *gormRepo) CreateApprovalRequest(ctx context.Context, model *gorm_model.ApprovalRequest) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) GetApprovalRequestByID(ctx context.Context, id string) (*gorm_model.ApprovalRequest, error) {
	var req gorm_model.ApprovalRequest
	err := r.db.WithContext(ctx).
		Preload("RequestedByAdmin").
		Preload("ReviewedBySuperadmin").
		Where("id = ?", id).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *gormRepo) FetchApprovalRequests(ctx context.Context, options gorm_model.ApprovalRequestFilter) ([]gorm_model.ApprovalRequest, error) {
	var requests []gorm_model.ApprovalRequest
	q := r.db.WithContext(ctx).
		Model(&gorm_model.ApprovalRequest{}).
		Preload("RequestedByAdmin").
		Preload("ReviewedBySuperadmin")
	options.Query(q)

	if err := q.Find(&requests).Error; err != nil {
		logrus.Error("FetchApprovalRequests error:", err)
		return nil, err
	}
	return requests, nil
}

func (r *gormRepo) CountApprovalRequests(ctx context.Context, options gorm_model.ApprovalRequestFilter) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&gorm_model.ApprovalRequest{})

	countFilter := options
	countFilter.Limit = nil
	countFilter.Offset = nil
	countFilter.Query(q)

	if err := q.Count(&count).Error; err != nil {
		logrus.Error("CountApprovalRequests error:", err)
		return 0, err
	}
	return count, nil
}

func (r *gormRepo) UpdateApprovalRequest(ctx context.Context, model *gorm_model.ApprovalRequest) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) GetPendingApprovalByRecord(ctx context.Context, tableName, recordID string) (*gorm_model.ApprovalRequest, error) {
	var req gorm_model.ApprovalRequest
	err := r.db.WithContext(ctx).
		Where("table_name = ? AND record_id = ? AND status = 'PENDING'", tableName, recordID).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *gormRepo) DeleteApprovalRequest(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&gorm_model.ApprovalRequest{}, "id = ?", id).Error
}
