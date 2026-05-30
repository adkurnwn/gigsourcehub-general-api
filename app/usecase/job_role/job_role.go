package usecase_job_role

import (
	"context"
	"net/http"
	"strconv"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/google/uuid"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
	var roles []gorm_model.JobRole
	dbFetch := u.gormDbRepo.GetDB().WithContext(ctx).Preload("Sector")
	filter.Query(dbFetch)

	if err := dbFetch.Find(&roles).Error; err != nil {
		logrus.Error("JobRole fetch error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Job Role data")
	}

	// Parse database response
	var results []interface{}
	for _, role := range roles {
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

	var role gorm_model.JobRole
	dbFetch := u.gormDbRepo.GetDB().WithContext(ctx).Preload("Sector")
	filter := gorm_model.JobRoleFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	}
	filter.Query(dbFetch)

	if err := dbFetch.First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Role not found")
		}
		logrus.Error("JobRole fetch data error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Role")
	}

	return response.Success(role.ToJobRoleResp())
}

func (u *appUsecase) Create(ctx context.Context, req request_model.CreateJobRoleRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Verify sector is active
	var sector gorm_model.Sector
	if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&sector, "id = ?", req.SectorID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusBadRequest, "Sector not found")
		}
		logrus.Error("JobRole Create sector verify error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to verify sector status")
	}
	if !sector.IsActive {
		return response.Error(http.StatusBadRequest, "Cannot reference an inactive sector")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Initializing new JobRole instance
	newRole := gorm_model.JobRole{
		ID:       uuid.New().String(),
		SectorID: req.SectorID,
		Name:     req.Name,
		IsActive: isActive,
	}

	if err := u.gormDbRepo.CreateJobRole(ctx, &newRole); err != nil {
		logrus.Error("JobRole Create error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Job Role", req.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create Job Role")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Job Role", req.Name, req, true)

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

	// Verify sector is active
	var sector gorm_model.Sector
	if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&sector, "id = ?", req.SectorID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusBadRequest, "Sector not found")
		}
		logrus.Error("JobRole Update sector verify error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to verify sector status")
	}
	if !sector.IsActive {
		return response.Error(http.StatusBadRequest, "Cannot reference an inactive sector")
	}

	// Deactivation safety check
	if existingRole.IsActive && req.IsActive != nil && !*req.IsActive {
		var assignedCount int64
		if err := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.User{}).Where("assigned_role_id = ?", id).Count(&assignedCount).Error; err != nil {
			logrus.Error("JobRole deactivation check error: ", err)
			return response.Error(http.StatusInternalServerError, "Failed to verify job role references")
		}
		var m2mCount int64
		if err := u.gormDbRepo.GetDB().WithContext(ctx).Table("user_has_job_roles").Where("job_role_id = ?", id).Count(&m2mCount).Error; err != nil {
			logrus.Error("JobRole deactivation check m2m error: ", err)
			return response.Error(http.StatusInternalServerError, "Failed to verify job role references")
		}
		var subrequestCount int64
		if err := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.Subrequest{}).Where("job_role_id = ?", id).Count(&subrequestCount).Error; err != nil {
			logrus.Error("JobRole deactivation check error: ", err)
			return response.Error(http.StatusInternalServerError, "Failed to verify job role references")
		}
		if assignedCount > 0 || m2mCount > 0 || subrequestCount > 0 {
			return response.Error(http.StatusBadRequest, "Cannot deactivate job role because it is currently referenced by one or more user profiles or subrequests")
		}
	}

	// Overwrite modifiable components
	existingRole.Name = req.Name
	existingRole.SectorID = req.SectorID
	if req.IsActive != nil {
		existingRole.IsActive = *req.IsActive
	}

	// Write modifications to DB
	if err := u.gormDbRepo.UpdateJobRole(ctx, &existingRole); err != nil {
		logrus.Error("JobRole Update error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Job Role", existingRole.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to update Job Role")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Job Role", existingRole.Name, req, true)

	return response.Success(existingRole.ToJobRoleResp())
}

func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Safety Check 1: Check if assigned to any user as AssignedRole
	var assignedCount int64
	if err := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.User{}).Where("assigned_role_id = ?", id).Count(&assignedCount).Error; err != nil {
		logrus.Error("JobRole delete check error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to check Job Role usage")
	}

	if assignedCount > 0 {
		return response.Error(http.StatusBadRequest, "Tidak bisa menghapus posisi karena sedang digunakan oleh user sebagai role utama")
	}

	// Safety Check 2: Check many-to-many user_has_job_roles
	var m2mCount int64
	if err := u.gormDbRepo.GetDB().WithContext(ctx).Table("user_has_job_roles").Where("job_role_id = ?", id).Count(&m2mCount).Error; err != nil {
		logrus.Error("JobRole delete check m2m error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to check Job Role usage in relations")
	}

	if m2mCount > 0 {
		return response.Error(http.StatusBadRequest, "Tidak bisa menghapus posisi karena sedang dipilih oleh kandidat")
	}

	if err := u.gormDbRepo.DeleteJobRole(ctx, id); err != nil {
		logrus.Error("JobRole Delete error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Job Role", id, nil, false)
		return response.Error(http.StatusInternalServerError, "Failed to delete Job Role")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Job Role", id, nil, true)

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
