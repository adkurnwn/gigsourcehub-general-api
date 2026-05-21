package usecase_interview_stage

import (
	"context"
	"net/http"
	"strconv"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (u *appUsecase) FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.InterviewStageFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	limitPtr := &limit
	offset := (page - 1) * limit
	filter.Limit = limitPtr
	filter.Offset = &offset
	if len(filter.Sorts) == 0 {
		filter.Sorts = []map[string]string{{"created_at": "DESC"}}
	}

	countFilter := filter
	countFilter.Limit = nil
	countFilter.Offset = nil

	dbCount := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.InterviewStage{})
	countFilter.Query(dbCount)
	var total int64
	if err := dbCount.Count(&total).Error; err != nil {
		logrus.Error("InterviewStage count error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to count Interview Stage records")
	}

	rows, err := u.gormDbRepo.FetchInterviewStage(ctx, filter)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Interview Stage data")
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var interviewStage gorm_model.InterviewStage
		if err := u.gormDbRepo.StructScan(rows, &interviewStage); err != nil {
			logrus.Error("InterviewStage map error:", err)
			continue
		}
		results = append(results, interviewStage.ToInterviewStageResp())
	}

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

	data, err := u.gormDbRepo.FetchInterviewStage(ctx, gorm_model.InterviewStageFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Interview Stage")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "Interview Stage not found")
	}

	var interviewStage gorm_model.InterviewStage
	if err := u.gormDbRepo.StructScan(data, &interviewStage); err != nil {
		logrus.Error("InterviewStage struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize Interview Stage data")
	}

	return response.Success(interviewStage.ToInterviewStageResp())
}

func (u *appUsecase) Create(ctx context.Context, req request_model.CreateInterviewStageRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	newInterviewStage := gorm_model.InterviewStage{
		ID:   uuid.New().String(),
		Name: req.Name,
	}

	if err := u.gormDbRepo.CreateInterviewStage(ctx, &newInterviewStage); err != nil {
		logrus.Error("InterviewStage Create error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Interview Stage", req.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create Interview Stage")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Interview Stage", req.Name, req, true)

	return response.Success(newInterviewStage.ToInterviewStageResp())
}

func (u *appUsecase) Update(ctx context.Context, id string, req request_model.UpdateInterviewStageRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	data, err := u.gormDbRepo.FetchInterviewStage(ctx, gorm_model.InterviewStageFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch existing Interview Stage")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "Interview Stage not found")
	}

	var existingInterviewStage gorm_model.InterviewStage
	if err := u.gormDbRepo.StructScan(data, &existingInterviewStage); err != nil {
		logrus.Error("InterviewStage struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize Interview Stage data")
	}

	existingInterviewStage.Name = req.Name

	if err := u.gormDbRepo.UpdateInterviewStage(ctx, &existingInterviewStage); err != nil {
		logrus.Error("InterviewStage Update error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Interview Stage", existingInterviewStage.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to update Interview Stage")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Interview Stage", existingInterviewStage.Name, req, true)

	return response.Success(existingInterviewStage.ToInterviewStageResp())
}

func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.gormDbRepo.DeleteInterviewStage(ctx, id); err != nil {
		logrus.Error("InterviewStage Delete error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Interview Stage", id, nil, false)
		return response.Error(http.StatusInternalServerError, "Failed to delete Interview Stage")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Interview Stage", id, nil, true)

	return response.Success(nil)
}