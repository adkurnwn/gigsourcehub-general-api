package usecase_job_title

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
	"gorm.io/gorm"
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
	var titles []gorm_model.JobTitle
	dbFetch := u.gormDbRepo.GetDB().WithContext(ctx).Preload("Sector")
	filter.Query(dbFetch)

	if err := dbFetch.Find(&titles).Error; err != nil {
		logrus.Error("JobTitle fetch error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Job Title data")
	}

	// Parse database response
	var results []interface{}
	for _, title := range titles {
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

	var title gorm_model.JobTitle
	dbFetch := u.gormDbRepo.GetDB().WithContext(ctx).Preload("Sector")
	filter := gorm_model.JobTitleFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	}
	filter.Query(dbFetch)

	if err := dbFetch.First(&title).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Title not found")
		}
		logrus.Error("JobTitle fetch data error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Title")
	}

	return response.Success(title.ToJobTitleResp())
}

func (u *appUsecase) Create(ctx context.Context, req request_model.CreateJobTitleRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Verify sector is active
	var sector gorm_model.Sector
	if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&sector, "id = ?", req.SectorID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusBadRequest, "Sector not found")
		}
		logrus.Error("JobTitle Create sector verify error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to verify sector status")
	}
	if !sector.IsActive {
		return response.Error(http.StatusBadRequest, "Cannot reference an inactive sector")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Initializing new JobTitle instance
	newTitle := gorm_model.JobTitle{
		ID:       uuid.New().String(),
		SectorID: req.SectorID,
		Name:     req.Name,
		IsActive: isActive,
	}

	if err := u.gormDbRepo.CreateJobTitle(ctx, &newTitle); err != nil {
		logrus.Error("JobTitle Create error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Job Title", req.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create Job Title")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Job Title", req.Name, req, true)

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

	// Verify sector is active
	var sector gorm_model.Sector
	if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&sector, "id = ?", req.SectorID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusBadRequest, "Sector not found")
		}
		logrus.Error("JobTitle Update sector verify error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to verify sector status")
	}
	if !sector.IsActive {
		return response.Error(http.StatusBadRequest, "Cannot reference an inactive sector")
	}

	// Deactivation safety check
	if existingTitle.IsActive && req.IsActive != nil && !*req.IsActive {
		var assignedCount int64
		if err := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.User{}).Where("job_title_id = ?", id).Count(&assignedCount).Error; err != nil {
			logrus.Error("JobTitle deactivation check error: ", err)
			return response.Error(http.StatusInternalServerError, "Failed to verify job title references")
		}
		if assignedCount > 0 {
			return response.Error(http.StatusBadRequest, "Cannot deactivate job title because it is currently referenced by one or more user profiles")
		}
	}

	// Overwrite modifiable components
	existingTitle.Name = req.Name
	existingTitle.SectorID = req.SectorID
	if req.IsActive != nil {
		existingTitle.IsActive = *req.IsActive
	}

	// Write modifications to DB
	if err := u.gormDbRepo.UpdateJobTitle(ctx, &existingTitle); err != nil {
		logrus.Error("JobTitle Update error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Job Title", existingTitle.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to update Job Title")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Job Title", existingTitle.Name, req, true)

	return response.Success(existingTitle.ToJobTitleResp())
}

func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Safety Check: Check if assigned to any user
	var assignedCount int64
	if err := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.User{}).Where("job_title_id = ?", id).Count(&assignedCount).Error; err != nil {
		logrus.Error("JobTitle delete check error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to check Job Title usage")
	}

	if assignedCount > 0 {
		return response.Error(http.StatusBadRequest, "Tidak bisa menghapus jabatan karena sedang digunakan oleh user")
	}

	if err := u.gormDbRepo.DeleteJobTitle(ctx, id); err != nil {
		logrus.Error("JobTitle Delete error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Job Title", id, nil, false)
		return response.Error(http.StatusInternalServerError, "Failed to delete Job Title")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Job Title", id, nil, true)

	return response.Success(nil)
}
