package usecase_job_title

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

func (u *appUsecase) FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.JobTitleFilter) response.Base {
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

	dbCount := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.JobTitle{})
	countFilter.Query(dbCount)
	var total int64
	if err := dbCount.Count(&total).Error; err != nil {
		logrus.Error("JobTitle count error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to count Job Title records")
	}

	// Query Database for Paginated Rows
	rows, err := u.gormDbRepo.FetchJobTitle(ctx, filter)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Job Title data")
	}
	defer rows.Close()

	// Parse database response cursors
	var results []interface{}
	for rows.Next() {
		var title gorm_model.JobTitle
		if err := u.gormDbRepo.StructScan(rows, &title); err != nil {
			logrus.Error("JobTitle map error:", err)
			continue
		}
		results = append(results, title.ToJobTitleResp())
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

	data, err := u.gormDbRepo.FetchJobTitle(ctx, gorm_model.JobTitleFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Title")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "Title not found")
	}

	var title gorm_model.JobTitle
	if err := u.gormDbRepo.StructScan(data, &title); err != nil {
		logrus.Error("Title struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize Title data")
	}

	return response.Success(title.ToJobTitleResp())
}

func (u *appUsecase) Create(ctx context.Context, req request_model.CreateJobTitleRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Initializing new JobTitle instance
	newTitle := gorm_model.JobTitle{
		ID:       uuid.New().String(),
		SectorID: req.SectorID,
		Name:     req.Name,
	}

	if err := u.gormDbRepo.CreateJobTitle(ctx, &newTitle); err != nil {
		logrus.Error("JobTitle Create error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to create Job Title")
	}

	return response.Success(newTitle.ToJobTitleResp())
}

func (u *appUsecase) Update(ctx context.Context, id string, req request_model.UpdateJobTitleRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Locate existing record
	data, err := u.gormDbRepo.FetchJobTitle(ctx, gorm_model.JobTitleFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch existing Title")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "Title not found")
	}

	var existingTitle gorm_model.JobTitle
	if err := u.gormDbRepo.StructScan(data, &existingTitle); err != nil {
		logrus.Error("Title struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize Title data")
	}

	// Overwrite modifiable components
	existingTitle.Name = req.Name
	existingTitle.SectorID = req.SectorID

	// Write modifications to DB
	if err := u.gormDbRepo.UpdateJobTitle(ctx, &existingTitle); err != nil {
		logrus.Error("JobTitle Update error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to update Job Title")
	}

	return response.Success(existingTitle.ToJobTitleResp())
}

func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.gormDbRepo.DeleteJobTitle(ctx, id); err != nil {
		logrus.Error("JobTitle Delete error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to delete Job Title")
	}

	return response.Success(nil)
}
