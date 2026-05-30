package usecase_job_vacancy

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// parseDateString parses a nullable date string "YYYY-MM-DD" to *time.Time.
func parseDateString(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	return &t, nil
}

// validateStatus ensures the status string is one of the allowed enum values.
func validateStatus(s *string) bool {
	if s == nil {
		return true
	}
	switch *s {
	case "DRAFT", "ARCHIVED", "PUBLISHED":
		return true
	}
	return false
}

// validateSchema ensures the schema string is one of the allowed enum values.
func validateSchema(s *string) bool {
	if s == nil {
		return true
	}
	switch *s {
	case "ONSITE", "REMOTE", "HYBRID":
		return true
	}
	return false
}

// FetchAll — CMS list with pagination, for Admin & Superadmin.
func (u *appUsecase) FetchAll(ctx context.Context, page, limit int64, filter gorm_model.JobVacancyFilter) response.Base {
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

	// Count total
	countFilter := filter
	countFilter.Limit = nil
	countFilter.Offset = nil
	dbCount := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.JobVacancy{})
	countFilter.Query(dbCount)
	var total int64
	if err := dbCount.Count(&total).Error; err != nil {
		logrus.Error("JobVacancy count error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to count Job Vacancy records")
	}

	// Fetch paginated
	var vacancies []gorm_model.JobVacancy
	dbFetch := u.gormDbRepo.GetDB().WithContext(ctx).
		Preload("Subrequest").
		Preload("Subrequest.Request").
		Preload("Subrequest.JobRole").
		Preload("Subrequest.JobRole.Sector")
	filter.Query(dbFetch)
	if err := dbFetch.Find(&vacancies).Error; err != nil {
		logrus.Error("JobVacancy fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Job Vacancy data")
	}

	var results []interface{}
	for _, v := range vacancies {
		results = append(results, v.ToJobVacancyResp())
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

// FetchData — CMS get by ID.
func (u *appUsecase) FetchData(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	vacancy, err := u.gormDbRepo.GetJobVacancyByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Job Vacancy not found")
		}
		logrus.Error("JobVacancy FetchData error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Job Vacancy")
	}

	return response.Success(vacancy.ToJobVacancyResp())
}

// Create — CMS create, for Admin & Superadmin.
func (u *appUsecase) Create(ctx context.Context, req request_model.CreateJobVacancyRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Validate enums
	if !validateStatus(req.Status) {
		return response.Error(http.StatusBadRequest, "Invalid status value. Allowed: DRAFT, ARCHIVED, PUBLISHED")
	}
	if !validateSchema(req.Schema) {
		return response.Error(http.StatusBadRequest, "Invalid schema value. Allowed: ONSITE, REMOTE, HYBRID")
	}

	// Parse dates
	takedownDate, err := parseDateString(req.TakedownDate)
	if err != nil {
		return response.Error(http.StatusBadRequest, "Invalid takedown_date: "+err.Error())
	}
	fulfillmentDate, err := parseDateString(req.FulfillmentDate)
	if err != nil {
		return response.Error(http.StatusBadRequest, "Invalid fulfillment_date: "+err.Error())
	}

	newVacancy := gorm_model.JobVacancy{
		ID:              uuid.New().String(),
		SubrequestID:    req.SubrequestID,
		Name:            req.Name,
		TakedownDate:    takedownDate,
		FulfillmentDate: fulfillmentDate,
		Schema:          req.Schema,
		Status:          req.Status,
		Description:     req.Description,
		Overview:        req.Overview,
	}

	if err := u.gormDbRepo.CreateJobVacancy(ctx, &newVacancy); err != nil {
		logrus.Error("JobVacancy Create error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Job Vacancy", req.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create Job Vacancy")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Job Vacancy", req.Name, req, true)
	return response.Success(newVacancy.ToJobVacancyResp())
}

// Update — CMS update, for Admin & Superadmin.
func (u *appUsecase) Update(ctx context.Context, id string, req request_model.UpdateJobVacancyRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Validate enums
	if !validateStatus(req.Status) {
		return response.Error(http.StatusBadRequest, "Invalid status value. Allowed: DRAFT, ARCHIVED, PUBLISHED")
	}
	if !validateSchema(req.Schema) {
		return response.Error(http.StatusBadRequest, "Invalid schema value. Allowed: ONSITE, REMOTE, HYBRID")
	}

	// Fetch existing
	existing, err := u.gormDbRepo.GetJobVacancyByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Job Vacancy not found")
		}
		logrus.Error("JobVacancy Update fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Job Vacancy")
	}

	// Parse dates
	takedownDate, err := parseDateString(req.TakedownDate)
	if err != nil {
		return response.Error(http.StatusBadRequest, "Invalid takedown_date: "+err.Error())
	}
	fulfillmentDate, err := parseDateString(req.FulfillmentDate)
	if err != nil {
		return response.Error(http.StatusBadRequest, "Invalid fulfillment_date: "+err.Error())
	}

	// Overwrite fields
	existing.SubrequestID = req.SubrequestID
	existing.Name = req.Name
	existing.TakedownDate = takedownDate
	existing.FulfillmentDate = fulfillmentDate
	existing.Schema = req.Schema
	existing.Status = req.Status
	existing.Description = req.Description
	existing.Overview = req.Overview

	if err := u.gormDbRepo.UpdateJobVacancy(ctx, existing); err != nil {
		logrus.Error("JobVacancy Update error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Job Vacancy", existing.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to update Job Vacancy")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Job Vacancy", existing.Name, req, true)
	return response.Success(existing.ToJobVacancyResp())
}

// Delete — CMS soft delete, for Admin & Superadmin.
func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Check existence first
	existing, err := u.gormDbRepo.GetJobVacancyByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Job Vacancy not found")
		}
		logrus.Error("JobVacancy Delete fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Job Vacancy")
	}

	if err := u.gormDbRepo.DeleteJobVacancy(ctx, id); err != nil {
		logrus.Error("JobVacancy Delete error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Job Vacancy", existing.Name, nil, false)
		return response.Error(http.StatusInternalServerError, "Failed to delete Job Vacancy")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "Job Vacancy", existing.Name, nil, true)
	return response.Success(nil)
}

// FetchPublic — Public list, tanpa auth, hanya PUBLISHED & belum takedown.
func (u *appUsecase) FetchPublic(ctx context.Context, page, limit int64, filter gorm_model.JobVacancyFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Force public filter
	filter.OnlyPublicValid = true

	limitPtr := &limit
	offset := (page - 1) * limit
	filter.Limit = limitPtr
	filter.Offset = &offset
	if len(filter.Sorts) == 0 {
		filter.Sorts = []map[string]string{
			{"created_at": "DESC"},
		}
	}

	// Count total
	countFilter := filter
	countFilter.Limit = nil
	countFilter.Offset = nil
	dbCount := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.JobVacancy{})
	countFilter.Query(dbCount)
	var total int64
	if err := dbCount.Count(&total).Error; err != nil {
		logrus.Error("JobVacancy FetchPublic count error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to count public Job Vacancy records")
	}

	// Fetch paginated
	var vacancies []gorm_model.JobVacancy
	dbFetch := u.gormDbRepo.GetDB().WithContext(ctx).
		Preload("Subrequest").
		Preload("Subrequest.Request").
		Preload("Subrequest.JobRole").
		Preload("Subrequest.JobRole.Sector")
	filter.Query(dbFetch)
	if err := dbFetch.Find(&vacancies).Error; err != nil {
		logrus.Error("JobVacancy FetchPublic fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch public Job Vacancy data")
	}

	var results []interface{}
	for _, v := range vacancies {
		results = append(results, v.ToJobVacancyPublicResp())
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

// FetchPublicByID — Public get by ID, hanya jika status valid & belum takedown.
func (u *appUsecase) FetchPublicByID(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	vacancy, err := u.gormDbRepo.GetJobVacancyByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Job Vacancy not found")
		}
		logrus.Error("JobVacancy FetchPublicByID error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Job Vacancy")
	}

	// Enforce public visibility rules
	if vacancy.Status == nil || *vacancy.Status != "PUBLISHED" {
		return response.Error(http.StatusNotFound, "Job Vacancy not found")
	}
	if vacancy.TakedownDate != nil && vacancy.TakedownDate.Before(time.Now().Truncate(24*time.Hour)) {
		return response.Error(http.StatusNotFound, "Job Vacancy not found")
	}

	return response.Success(vacancy.ToJobVacancyPublicResp())
}
