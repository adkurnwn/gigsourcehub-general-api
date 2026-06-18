package gormrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func applyRequestFilter(q *gorm.DB, filter gorm_model.RequestFilter) *gorm.DB {
	if filter.Status != nil && *filter.Status != "" {
		q = q.Where("requests.status = ?", *filter.Status)
	}
	if filter.Urgency != nil && *filter.Urgency != "" {
		q = q.Where("requests.urgency = ?", *filter.Urgency)
	}
	if filter.Search != nil && *filter.Search != "" {
		q = q.Where("requests.project_name ILIKE ?", "%"+*filter.Search+"%")
	}
	if filter.AdminUserID != nil && *filter.AdminUserID != "" {
		q = q.Where("requests.admin_user_id = ?", *filter.AdminUserID)
	}
	if filter.ProposedBy != nil && *filter.ProposedBy != "" {
		q = q.Joins("LEFT JOIN users AS employee ON employee.id = requests.employee_user_id").
			Where("employee.name ILIKE ?", "%"+*filter.ProposedBy+"%")
	}
	if filter.AdminName != nil && *filter.AdminName != "" {
		q = q.Joins("LEFT JOIN users AS admin ON admin.id = requests.admin_user_id").
			Where("admin.name ILIKE ?", "%"+*filter.AdminName+"%")
	}
	return q
}

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
		Preload("EmployeeUser").
		Preload("Subrequests").
		Preload("Subrequests.JobRole").
		Preload("Subrequests.JobRole.Sector").
		Where("employee_user_id = ?", employeeID).
		Order("requests.created_at DESC").
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

func (r *gormRepo) FetchRequestsByAdmin(ctx context.Context, filter gorm_model.RequestFilter, limit, offset int64) (*sql.Rows, error) {
	q := r.db.WithContext(ctx).Model(&gorm_model.Request{}).
		Preload("AdminUser").
		Preload("EmployeeUser").
		Preload("Subrequests").
		Preload("Subrequests.JobRole").
		Preload("Subrequests.JobRole.Sector").
		Order("requests.created_at DESC").
		Limit(int(limit)).Offset(int(offset))

	q = applyRequestFilter(q, filter)

	rows, err := q.Rows()
	if err != nil {
		logrus.Errorf("FetchRequestsByAdmin DB Error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

func (r *gormRepo) CountRequestsByAdmin(ctx context.Context, filter gorm_model.RequestFilter) (int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&gorm_model.Request{})
	q = applyRequestFilter(q, filter)
	if err := q.Count(&total).Error; err != nil {
		logrus.Errorf("CountRequestsByAdmin DB Error: %v\n", err)
		return 0, err
	}
	return total, nil
}

func (r *gormRepo) GetRequestByID(ctx context.Context, id string) (*gorm_model.Request, error) {
	var request gorm_model.Request
	err := r.db.WithContext(ctx).
		Preload("AdminUser").
		Preload("EmployeeUser").
		Preload("Subrequests").
		Preload("Subrequests.JobRole").
		Preload("Subrequests.JobRole.Sector").
		Where("requests.id = ?", id).
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

func (r *gormRepo) UpdateRequestByAdmin(ctx context.Context, model *gorm_model.Request) error {
	if err := r.db.WithContext(ctx).Model(&gorm_model.Request{}).Where("id = ?", model.ID).Updates(model).Error; err != nil {
		logrus.Errorf("UpdateRequestByAdmin DB Error: %v\n", err)
		return err
	}
	return nil
}

func (r *gormRepo) UpdateRequestAdminUser(ctx context.Context, requestID string, adminUserID string) error {
	if err := r.db.WithContext(ctx).Model(&gorm_model.Request{}).
		Where("id = ?", requestID).
		Update("admin_user_id", adminUserID).Error; err != nil {
		logrus.Errorf("UpdateRequestAdminUser DB Error: %v\n", err)
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

func (r *gormRepo) AssignCandidateToSubrequest(ctx context.Context, model *gorm_model.SubrequestCandidate, recruitmentStatusID string, requestID string, markRequestProcessing bool) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existingCount int64
		if err := tx.Model(&gorm_model.SubrequestCandidate{}).
			Where("subrequest_id = ? AND candidate_user_id = ? AND deleted_at IS NULL", model.SubrequestID, model.CandidateUserID).
			Count(&existingCount).Error; err != nil {
			return err
		}
		if existingCount > 0 {
			return errors.New("candidate already assigned to this subrequest")
		}

		if err := tx.Create(model).Error; err != nil {
			return err
		}

		if markRequestProcessing {
			if err := tx.Model(&gorm_model.Request{}).
				Where("id = ?", requestID).
				Update("status", "PROCESSING").Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&gorm_model.User{}).
			Where("id = ?", model.CandidateUserID).
			Update("recruitment_status_id", recruitmentStatusID).Error; err != nil {
			return err
		}

		changerID := helpers.GetActorID(ctx)
		if changerID != "" {
			history := &gorm_model.CandidateStatusHistory{
				CandidateUserID:     model.CandidateUserID,
				RecruitmentStatusID: &recruitmentStatusID,
				SubrequestID:        &model.SubrequestID,
				ChangedByUserID:     changerID,
			}
			if err := tx.Create(history).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logrus.Errorf("AssignCandidateToSubrequest DB Error: %v\n", err)
	}
	return err
}

func (r *gormRepo) CountActiveSubrequestCandidatesByCandidateID(ctx context.Context, candidateID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&gorm_model.SubrequestCandidate{}).
		Where("candidate_user_id = ? AND deleted_at IS NULL", candidateID).
		Count(&total).Error
	if err != nil {
		logrus.Errorf("CountActiveSubrequestCandidatesByCandidateID DB Error: %v", err)
		return 0, err
	}
	return total, nil
}

func (r *gormRepo) SoftDeleteSubrequestCandidatesByCandidateID(ctx context.Context, candidateID string) error {
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&gorm_model.SubrequestCandidate{}).
		Where("candidate_user_id = ? AND deleted_at IS NULL", candidateID).
		Update("deleted_at", now).Error
	if err != nil {
		logrus.Errorf("SoftDeleteSubrequestCandidatesByCandidateID DB Error: %v", err)
	}
	return err
}

func (r *gormRepo) StopOnboardingByCandidateID(ctx context.Context, candidateID string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var onboardHistoryID string
		if err := tx.Model(&gorm_model.OnboardHistory{}).
			Where("candidate_user_id = ? AND deleted_at IS NULL", candidateID).
			Order("created_at DESC").
			Limit(1).
			Pluck("id", &onboardHistoryID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if onboardHistoryID == "" {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Model(&gorm_model.OnboardHistory{}).
			Where("id = ?", onboardHistoryID).
			Update("is_stopped", true).Error; err != nil {
			return err
		}

		now := time.Now()
		if err := tx.Model(&gorm_model.SubrequestCandidate{}).
			Where("candidate_user_id = ? AND deleted_at IS NULL", candidateID).
			Update("deleted_at", now).Error; err != nil {
			return err
		}

		if err := tx.Model(&gorm_model.User{}).
			Where("id = ?", candidateID).
			Update("recruitment_status_id", nil).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logrus.Errorf("StopOnboardingByCandidateID DB Error: %v", err)
	}
	return err
}

func (r *gormRepo) CancelRecruitmentByCandidateID(ctx context.Context, candidateID string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var subReqID *string
		var latestSubrequestID string
		if err := tx.Model(&gorm_model.SubrequestCandidate{}).
			Where("candidate_user_id = ? AND deleted_at IS NULL", candidateID).
			Order("created_at DESC").
			Limit(1).
			Pluck("subrequest_id", &latestSubrequestID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if latestSubrequestID != "" {
			subReqID = &latestSubrequestID
		}

		now := time.Now()
		if err := tx.Model(&gorm_model.SubrequestCandidate{}).
			Where("candidate_user_id = ? AND deleted_at IS NULL", candidateID).
			Update("deleted_at", now).Error; err != nil {
			return err
		}

		if err := tx.Model(&gorm_model.User{}).
			Where("id = ?", candidateID).
			Update("recruitment_status_id", nil).Error; err != nil {
			return err
		}

		changerID := helpers.GetActorID(ctx)
		if changerID != "" {
			history := &gorm_model.CandidateStatusHistory{
				CandidateUserID:     candidateID,
				RecruitmentStatusID: nil,
				SubrequestID:        subReqID,
				ChangedByUserID:     changerID,
			}
			if err := tx.Create(history).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logrus.Errorf("CancelRecruitmentByCandidateID DB Error: %v", err)
	}
	return err
}

func (r *gormRepo) GetActiveSubrequestByCandidateID(ctx context.Context, candidateID string) (*gorm_model.ActiveSubrequestInfo, error) {
	var info gorm_model.ActiveSubrequestInfo
	err := r.db.WithContext(ctx).
		Table("subrequest_candidates").
		Select("subrequest_candidates.subrequest_id, subrequests.request_id, requests.project_name, job_roles.name as job_role, subrequest_candidates.created_at, subrequest_candidates.declined_reason").
		Joins("JOIN subrequests ON subrequests.id = subrequest_candidates.subrequest_id").
		Joins("JOIN requests ON requests.id = subrequests.request_id").
		Joins("LEFT JOIN job_roles ON job_roles.id = subrequests.job_role_id").
		Where("subrequest_candidates.candidate_user_id = ? AND subrequest_candidates.deleted_at IS NULL AND subrequests.deleted_at IS NULL AND requests.deleted_at IS NULL", candidateID).
		Order("subrequest_candidates.created_at DESC").
		Limit(1).
		Take(&info).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logrus.Errorf("GetActiveSubrequestByCandidateID DB Error: %v", err)
		return nil, err
	}

	return &info, nil
}
