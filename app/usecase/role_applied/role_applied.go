package usecase_role_applied

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

func (u *appUsecase) FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.RoleAppliedFilter) response.Base {
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

	dbCount := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.RoleApplied{})
	countFilter.Query(dbCount)
	var total int64
	if err := dbCount.Count(&total).Error; err != nil {
		logrus.Error("RoleApplied count error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to count Role Applied records")
	}

	// Query Database for Paginated Rows
	rows, err := u.gormDbRepo.FetchRoleApplied(ctx, filter)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Role Applied data")
	}
	defer rows.Close()

	// Parse database response cursors
	var results []interface{}
	for rows.Next() {
		var role gorm_model.RoleApplied
		if err := u.gormDbRepo.StructScan(rows, &role); err != nil {
			logrus.Error("RoleApplied map error:", err)
			continue
		}
		results = append(results, role.ToRoleAppliedResp())
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

	data, err := u.gormDbRepo.FetchRoleApplied(ctx, gorm_model.RoleAppliedFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch Role")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "Role not found")
	}

	var role gorm_model.RoleApplied
	if err := u.gormDbRepo.StructScan(data, &role); err != nil {
		logrus.Error("Role struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize Role data")
	}

	return response.Success(role.ToRoleAppliedResp())
}

func (u *appUsecase) Create(ctx context.Context, req request_model.CreateRoleAppliedRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Initializing new RoleApplied instance
	newRole := gorm_model.RoleApplied{
		ID:       uuid.New().String(),
		SectorID: req.SectorID,
		Name:     req.Name,
	}

	if err := u.gormDbRepo.CreateRoleApplied(ctx, &newRole); err != nil {
		logrus.Error("RoleApplied Create error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to create Role Applied")
	}

	return response.Success(newRole.ToRoleAppliedResp())
}

func (u *appUsecase) Update(ctx context.Context, id string, req request_model.UpdateRoleAppliedRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Locate existing record
	data, err := u.gormDbRepo.FetchRoleApplied(ctx, gorm_model.RoleAppliedFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: id},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch existing Role")
	}
	defer data.Close()

	if !data.Next() {
		return response.Error(http.StatusNotFound, "Role not found")
	}

	var existingRole gorm_model.RoleApplied
	if err := u.gormDbRepo.StructScan(data, &existingRole); err != nil {
		logrus.Error("Role struct map error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to serialize Role data")
	}

	// Overwrite modifiable components
	existingRole.Name = req.Name
	existingRole.SectorID = req.SectorID

	// Write modifications to DB
	if err := u.gormDbRepo.UpdateRoleApplied(ctx, &existingRole); err != nil {
		logrus.Error("RoleApplied Update error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to update Role Applied")
	}

	return response.Success(existingRole.ToRoleAppliedResp())
}

func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.gormDbRepo.DeleteRoleApplied(ctx, id); err != nil {
		logrus.Error("RoleApplied Delete error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to delete Role Applied")
	}

	return response.Success(nil)
}
