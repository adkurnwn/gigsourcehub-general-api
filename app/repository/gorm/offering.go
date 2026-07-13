package gormrepo

import (
	"context"
	"errors"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (r *gormRepo) GetFinalizeSnapshotData(ctx context.Context, subrequestID string) (*gorm_model.FinalizeRecruitmentSnapshot, error) {
	var snapshot gorm_model.FinalizeRecruitmentSnapshot
	tx := r.db.WithContext(ctx).
		Table("subrequests").
		Select("subrequests.id AS subrequest_id, subrequests.request_id, subrequests.job_role_id, job_roles.name AS job_role_name, requests.project_name, requests.employee_user_id, users.name AS employee_name").
		Joins("JOIN requests ON requests.id = subrequests.request_id").
		Joins("JOIN users ON users.id = requests.employee_user_id").
		Joins("LEFT JOIN job_roles ON job_roles.id = subrequests.job_role_id").
		Where("subrequests.id = ? AND subrequests.deleted_at IS NULL AND requests.deleted_at IS NULL", subrequestID).
		Scan(&snapshot)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, errors.New("record not found")
	}

	return &snapshot, nil
}

func (r *gormRepo) FinalizeRecruitment(ctx context.Context, candidateID, subrequestID, requestID, acceptedStatusID string, startDate, endDate *time.Time, snapshotJSON string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		history := &gorm_model.OnboardHistory{
			CandidateUserID: candidateID,
			IsStopped:       false,
			StartDate:       startDate,
			EndDate:         endDate,
			Snapshot:        &snapshotJSON,
		}
		if err := tx.Create(history).Error; err != nil {
			logrus.Errorf("FinalizeRecruitment create onboard_history DB Error: %v", err)
			return err
		}

		if err := tx.Model(&gorm_model.Subrequest{}).
			Where("id = ?", subrequestID).
			Update("is_filled", true).Error; err != nil {
			logrus.Errorf("FinalizeRecruitment update subrequest is_filled DB Error: %v", err)
			return err
		}

		if err := tx.Model(&gorm_model.User{}).
			Where("id = ?", candidateID).
			Update("recruitment_status_id", acceptedStatusID).Error; err != nil {
			logrus.Errorf("FinalizeRecruitment update candidate status DB Error: %v", err)
			return err
		}

		changerID := helpers.GetActorID(ctx)
		if changerID != "" {
			history := &gorm_model.CandidateStatusHistory{
				CandidateUserID:     candidateID,
				RecruitmentStatusID: &acceptedStatusID,
				SubrequestID:        &subrequestID,
				ChangedByUserID:     changerID,
			}
			if err := tx.Create(history).Error; err != nil {
				logrus.Errorf("FinalizeRecruitment create status history DB Error: %v", err)
				return err
			}
		}

		// Fetch other candidates IDs to log their status change to nil
		var otherCandidateIDs []string
		tx.Model(&gorm_model.SubrequestCandidate{}).
			Where("subrequest_id = ? AND candidate_user_id <> ? AND deleted_at IS NULL", subrequestID, candidateID).
			Pluck("candidate_user_id", &otherCandidateIDs)

		subQuery := tx.Table("subrequest_candidates").
			Select("candidate_user_id").
			Where("subrequest_id = ? AND candidate_user_id <> ? AND deleted_at IS NULL", subrequestID, candidateID)

		if err := tx.Model(&gorm_model.User{}).
			Where("id IN (?)", subQuery).
			Update("recruitment_status_id", nil).Error; err != nil {
			logrus.Errorf("FinalizeRecruitment update other candidates status DB Error: %v", err)
			return err
		}

		if changerID != "" && len(otherCandidateIDs) > 0 {
			for _, otherID := range otherCandidateIDs {
				history := &gorm_model.CandidateStatusHistory{
					CandidateUserID:     otherID,
					RecruitmentStatusID: nil,
					SubrequestID:        &subrequestID,
					ChangedByUserID:     changerID,
				}
				if err := tx.Create(history).Error; err != nil {
					logrus.Errorf("FinalizeRecruitment create other status history DB Error: %v", err)
					return err
				}
			}
		}

		now := time.Now()
		if err := tx.Model(&gorm_model.SubrequestCandidate{}).
			Where("subrequest_id = ? AND candidate_user_id <> ? AND deleted_at IS NULL", subrequestID, candidateID).
			Update("deleted_at", now).Error; err != nil {
			logrus.Errorf("FinalizeRecruitment delete other candidates DB Error: %v", err)
			return err
		}

		var remainingUnfilled int64
		if err := tx.Model(&gorm_model.Subrequest{}).
			Where("request_id = ? AND deleted_at IS NULL AND is_filled = false", requestID).
			Count(&remainingUnfilled).Error; err != nil {
			logrus.Errorf("FinalizeRecruitment count remaining subrequests DB Error: %v", err)
			return err
		}

		if remainingUnfilled == 0 {
			today := helpers.GetCurrentTime()
			if err := tx.Model(&gorm_model.Request{}).
				Where("id = ?", requestID).
				Update("fulfillment_date", today).Error; err != nil {
				logrus.Errorf("FinalizeRecruitment update fulfillment_date DB Error: %v", err)
				return err
			}

			if err := tx.Model(&gorm_model.Request{}).
				Where("id = ? AND status = ?", requestID, "PROCESSING").
				Update("status", "DONE").Error; err != nil {
				logrus.Errorf("FinalizeRecruitment update request status DB Error: %v", err)
				return err
			}
		}

		return nil
	})
}
