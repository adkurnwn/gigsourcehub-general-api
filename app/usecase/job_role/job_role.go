package usecase_job_role

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

func (u *appUsecase) FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.JobRoleFilter) response.Base {
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

	dbCount := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.JobRole{})
	countFilter.Query(dbCount)
	var total int64
	if err := dbCount.Count(&total).Error; err != nil {
		logrus.Error("JobRole count error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to count Job Role records")
	}

	// Query Database for Paginated Rows
	rows, err := u.gormDbRepo.FetchJobRole(ctx, filter)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Job Role data")
	}
	defer rows.Close()

	// Parse database response cursors
	var results []interface{}
	for rows.Next() {
		var role gorm_model.JobRole
		if err := u.gormDbRepo.StructScan(rows, &role); err != nil {
			logrus.Error("JobRole map error:", err)
			continue
		}
		results = append(results, role.ToJobRoleResp())
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

	data, err := u.gormDbRepo.FetchJobRole(ctx, gorm_model.JobRoleFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Role")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "Role not found")
	}

	var role gorm_model.JobRole
	if err := u.gormDbRepo.StructScan(data, &role); err != nil {
		logrus.Error("Role struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize Role data")
	}

	return response.Success(role.ToJobRoleResp())
}

func (u *appUsecase) Create(ctx context.Context, req request_model.CreateJobRoleRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Initializing new JobRole instance
	newRole := gorm_model.JobRole{
		ID:       uuid.New().String(),
		SectorID: req.SectorID,
		Name:     req.Name,
	}

	if err := u.gormDbRepo.CreateJobRole(ctx, &newRole); err != nil {
		logrus.Error("JobRole Create error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to create Job Role")
	}

	return response.Success(newRole.ToJobRoleResp())
}

func (u *appUsecase) Update(ctx context.Context, id string, req request_model.UpdateJobRoleRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Locate existing record
	data, err := u.gormDbRepo.FetchJobRole(ctx, gorm_model.JobRoleFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch existing Role")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "Role not found")
	}

	var existingRole gorm_model.JobRole
	if err := u.gormDbRepo.StructScan(data, &existingRole); err != nil {
		logrus.Error("Role struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize Role data")
	}

	// Overwrite modifiable components
	existingRole.Name = req.Name
	existingRole.SectorID = req.SectorID

	// Write modifications to DB
	if err := u.gormDbRepo.UpdateJobRole(ctx, &existingRole); err != nil {
		logrus.Error("JobRole Update error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to update Job Role")
	}

	return response.Success(existingRole.ToJobRoleResp())
}

func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.gormDbRepo.DeleteJobRole(ctx, id); err != nil {
		logrus.Error("JobRole Delete error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to delete Job Role")
	}

	return response.Success(nil)
}

func (u *appUsecase) FetchSystemRoles(ctx context.Context) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	rows, err := u.gormDbRepo.FetchSystemRole(ctx, gorm_model.SystemRoleFilter{})
	if err != nil {
		logrus.Error("FetchSystemRoles error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch system roles")
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var role gorm_model.SystemRole
		if err := u.gormDbRepo.StructScan(rows, &role); err != nil {
			logrus.Error("SystemRole map error:", err)
			continue
		}
		results = append(results, role.ToSystemRoleResp())
	}

	return response.Success(results)
}
