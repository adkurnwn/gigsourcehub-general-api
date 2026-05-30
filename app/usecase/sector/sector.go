package usecase_sector

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

func (u *appUsecase) FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.SectorFilter) response.Base {
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

	dbCount := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.Sector{})
	countFilter.Query(dbCount)
	var total int64
	if err := dbCount.Count(&total).Error; err != nil {
		logrus.Error("Sector count error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to count Sector records")
	}

	// Query Database for Paginated Rows
	rows, err := u.gormDbRepo.FetchSector(ctx, filter)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Sector data")
	}
	defer rows.Close()

	// Parse database response cursors
	var results []interface{}
	for rows.Next() {
		var sector gorm_model.Sector
		if err := u.gormDbRepo.StructScan(rows, &sector); err != nil {
			logrus.Error("Sector map error:", err)
			continue
		}
		results = append(results, sector.ToSectorResp())
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

	data, err := u.gormDbRepo.FetchSector(ctx, gorm_model.SectorFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Sector")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "Sector not found")
	}

	var sector gorm_model.Sector
	if err := u.gormDbRepo.StructScan(data, &sector); err != nil {
		logrus.Error("Sector struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize Sector data")
	}

	return response.Success(sector.ToSectorResp())
}

func (u *appUsecase) Create(ctx context.Context, req request_model.CreateSectorRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Initializing new Sector instance
	newSector := gorm_model.Sector{
		ID:       uuid.New().String(),
		Name:     req.Name,
	}

	if err := u.gormDbRepo.CreateSector(ctx, &newSector); err != nil {
		logrus.Error("Sector Create error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Sector", req.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create Sector")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Sector", req.Name, req, true)

	return response.Success(newSector.ToSectorResp())
}

func (u *appUsecase) Update(ctx context.Context, id string, req request_model.UpdateSectorRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Locate existing record
	data, err := u.gormDbRepo.FetchSector(ctx, gorm_model.SectorFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch existing Sector")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "Sector not found")
	}

	var existingSector gorm_model.Sector
	if err := u.gormDbRepo.StructScan(data, &existingSector); err != nil {
		logrus.Error("Sector struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize Sector data")
	}

	// Overwrite modifiable components
	if existingSector.IsActive && !req.IsActive {
		var jobRoleCount int64
		if err := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.JobRole{}).Where("sector_id = ?", id).Count(&jobRoleCount).Error; err != nil {
			logrus.Error("Sector deactivation check error: ", err)
			return response.Error(http.StatusInternalServerError, "Failed to verify sector references")
		}
		var jobTitleCount int64
		if err := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.JobTitle{}).Where("sector_id = ?", id).Count(&jobTitleCount).Error; err != nil {
			logrus.Error("Sector deactivation check error: ", err)
			return response.Error(http.StatusInternalServerError, "Failed to verify sector references")
		}
		if jobRoleCount > 0 || jobTitleCount > 0 {
			return response.Error(http.StatusBadRequest, "Cannot deactivate sector because it is currently referenced by one or more job roles or job titles")
		}
	}

	existingSector.Name = req.Name
	existingSector.IsActive = req.IsActive

	// Write modifications to DB
	if err := u.gormDbRepo.UpdateSector(ctx, &existingSector); err != nil {
		logrus.Error("Sector Update error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Sector", existingSector.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to update Sector")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Sector", existingSector.Name, req, true)

	return response.Success(existingSector.ToSectorResp())
}

func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	var jobRoleCount int64
	if err := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.JobRole{}).Where("sector_id = ?", id).Count(&jobRoleCount).Error; err != nil {
		logrus.Error("Sector delete check error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to verify sector references")
	}
	var jobTitleCount int64
	if err := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.JobTitle{}).Where("sector_id = ?", id).Count(&jobTitleCount).Error; err != nil {
		logrus.Error("Sector delete check error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to verify sector references")
	}
	if jobRoleCount > 0 || jobTitleCount > 0 {
		return response.Error(http.StatusBadRequest, "Cannot delete sector because it is currently referenced by one or more job roles or job titles")
	}

	if err := u.gormDbRepo.DeleteSector(ctx, id); err != nil {
		logrus.Error("Sector Delete error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Sector", id, nil, false)
		return response.Error(http.StatusInternalServerError, "Failed to delete Sector")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Sector", id, nil, true)

	return response.Success(nil)
}
