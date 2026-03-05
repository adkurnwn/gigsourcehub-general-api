package usecase_recruitment_status

import (
	"context"
	"net/http"
	"strconv"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (u *appUsecase) FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.RecruitmentStatusFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	limitPtr := &limit
	offset := (page - 1) * limit
	filter.Limit = limitPtr
	filter.Offset = &offset
	if len(filter.Sorts) == 0 {
		filter.Sorts = []map[string]string{
			{"created_at": "DESC"},
		}
	}

	// Calculate overall total database items
	countFilter := filter
	countFilter.Limit = nil
	countFilter.Offset = nil

	dbCount := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.RecruitmentStatus{})
	countFilter.Query(dbCount)
	var total int64
	if err := dbCount.Count(&total).Error; err != nil {
		logrus.Error("RecruitmentStatus count error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to count RecruitmentStatus records")
	}

	// Query Database for Paginated Rows
	rows, err := u.gormDbRepo.FetchRecruitmentStatus(ctx, filter)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch RecruitmentStatus data")
	}
	defer rows.Close()

	// Parse database response cursors
	var results []interface{}
	for rows.Next() {
		var recruitmentStatus gorm_model.RecruitmentStatus
		if err := u.gormDbRepo.StructScan(rows, &recruitmentStatus); err != nil {
			logrus.Error("RecruitmentStatus map error:", err)
			continue
		}
		results = append(results, recruitmentStatus.ToRecruitmentStatusResp())
	}

	// Assign Cursor token pointer for the next request block
	var nextCursor *string
	if offset+limit < total {
		nextStr := strconv.FormatInt(page+1, 10)
		nextCursor = &nextStr
	}

	return response.Success(response.List{
		List:   results,
		Limit:  limit,
		Page:   page,
		Total:  total,
		Cursor: nextCursor,
	})
}

func (u *appUsecase) FetchData(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	data, err := u.gormDbRepo.FetchRecruitmentStatus(ctx, gorm_model.RecruitmentStatusFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch RecruitmentStatus")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "RecruitmentStatus not found")
	}

	var recruitmentStatus gorm_model.RecruitmentStatus
	if err := u.gormDbRepo.StructScan(data, &recruitmentStatus); err != nil {
		logrus.Error("RecruitmentStatus struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize RecruitmentStatus data")
	}

	return response.Success(recruitmentStatus.ToRecruitmentStatusResp())
}

func (u *appUsecase) Create(ctx context.Context, req request_model.CreateRecruitmentStatusRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Initializing new RecruitmentStatus instance
	newRecruitmentStatus := gorm_model.RecruitmentStatus{
		ID:       uuid.New().String(),
		Name:     req.Name,
	}

	if err := u.gormDbRepo.CreateRecruitmentStatus(ctx, &newRecruitmentStatus); err != nil {
		logrus.Error("RecruitmentStatus Create error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to create RecruitmentStatus")
	}

	return response.Success(newRecruitmentStatus.ToRecruitmentStatusResp())
}

func (u *appUsecase) Update(ctx context.Context, id string, req request_model.UpdateRecruitmentStatusRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Locate existing record
	data, err := u.gormDbRepo.FetchRecruitmentStatus(ctx, gorm_model.RecruitmentStatusFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch existing RecruitmentStatus")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "RecruitmentStatus not found")
	}

	var existingRecruitmentStatus gorm_model.RecruitmentStatus
	if err := u.gormDbRepo.StructScan(data, &existingRecruitmentStatus); err != nil {
		logrus.Error("RecruitmentStatus struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize RecruitmentStatus data")
	}

	// Overwrite modifiable components
	existingRecruitmentStatus.Name = req.Name

	// Write modifications to DB
	if err := u.gormDbRepo.UpdateRecruitmentStatus(ctx, &existingRecruitmentStatus); err != nil {
		logrus.Error("RecruitmentStatus Update error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to update RecruitmentStatus")
	}

	return response.Success(existingRecruitmentStatus.ToRecruitmentStatusResp())
}

func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.gormDbRepo.DeleteRecruitmentStatus(ctx, id); err != nil {
		logrus.Error("RecruitmentStatus Delete error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to delete RecruitmentStatus")
	}

	return response.Success(nil)
}