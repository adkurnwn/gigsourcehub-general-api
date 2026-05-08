package gormrepo

import (
	"context"
	"database/sql"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (r *gormRepo) CreateRequest(ctx context.Context, model *gorm_model.Request) error {
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		logrus.Errorf("CreateRequest DB Error: %v\n", err)
		return err
	}
	return nil
}

func (r *gormRepo) FetchRequestsByEmployee(ctx context.Context, employeeID string, limit, offset int64) (*sql.Rows, error) {
	q := r.db.WithContext(ctx).Model(&gorm_model.Request{}).
		Preload("AdminUser").
		Preload("Subrequests").
		Preload("Subrequests.JobRole").
		Where("employee_user_id = ?", employeeID).
		Order("created_at DESC").
		Limit(int(limit)).Offset(int(offset))

	rows, err := q.Rows()
	if err != nil {
		logrus.Errorf("FetchRequestsByEmployee DB Error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

func (r *gormRepo) CountRequestsByEmployee(ctx context.Context, employeeID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&gorm_model.Request{}).
		Where("employee_user_id = ?", employeeID).Count(&total).Error
	if err != nil {
		logrus.Errorf("CountRequestsByEmployee DB Error: %v\n", err)
		return 0, err
	}
	return total, nil
}

func (r *gormRepo) FetchAllRequests(ctx context.Context, limit, offset int64) (*sql.Rows, error) {
	q := r.db.WithContext(ctx).Model(&gorm_model.Request{}).
		Preload("AdminUser").
		Preload("Subrequests").
		Preload("Subrequests.JobRole").
		Order("created_at DESC").
		Limit(int(limit)).Offset(int(offset))

	rows, err := q.Rows()
	if err != nil {
		logrus.Errorf("FetchAllRequests DB Error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

func (r *gormRepo) CountAllRequests(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&gorm_model.Request{}).Count(&total).Error
	if err != nil {
		logrus.Errorf("CountAllRequests DB Error: %v\n", err)
		return 0, err
	}
	return total, nil
}

func (r *gormRepo) GetRequestByID(ctx context.Context, id string) (*gorm_model.Request, error) {
	var request gorm_model.Request
	err := r.db.WithContext(ctx).
		Preload("AdminUser").
		Preload("Subrequests").
		Preload("Subrequests.JobRole").
		Where("id = ?", id).
		First(&request).Error
	if err != nil {
		logrus.Errorf("GetRequestByID DB Error: %v\n", err)
		return nil, err
	}
	return &request, nil
}

func (r *gormRepo) UpdateRequestByEmployee(ctx context.Context, model *gorm_model.Request) error {
	if err := r.db.WithContext(ctx).Model(&gorm_model.Request{}).Where("id = ?", model.ID).Updates(model).Error; err != nil {
		logrus.Errorf("UpdateRequestByEmployee DB Error: %v\n", err)
		return err
	}
	return nil
}

func (r *gormRepo) GetSubrequestByID(ctx context.Context, id string) (*gorm_model.Subrequest, error) {
	var subrequest gorm_model.Subrequest
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&subrequest).Error; err != nil {
		return nil, err
	}
	return &subrequest, nil
}

func (r *gormRepo) UpdateSubrequestByEmployee(ctx context.Context, model *gorm_model.Subrequest) error {
	if err := r.db.WithContext(ctx).Model(&gorm_model.Subrequest{}).Where("id = ?", model.ID).Updates(model).Error; err != nil {
		logrus.Errorf("UpdateSubrequestByEmployee DB Error: %v\n", err)
		return err
	}
	return nil
}

func (r *gormRepo) CreateSubrequestByEmployee(ctx context.Context, model *gorm_model.Subrequest) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Insert new subrequest natively
		if err := tx.Create(model).Error; err != nil {
			return err
		}

		// 2. Atomically increment the parent request's required_headcount
		if err := tx.Model(&gorm_model.Request{}).Where("id = ?", model.RequestID).UpdateColumn("required_headcount", gorm.Expr("required_headcount + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logrus.Errorf("CreateSubrequestByEmployee DB Error: %v\n", err)
	}
	return err
}
