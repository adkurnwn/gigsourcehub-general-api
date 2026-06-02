package usecase_member

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
)

func (u *appUsecase) FetchOnboardingByCandidate(ctx context.Context, candidateID string, page, limit int64, cursor string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if candidateID == "" {
		return response.Error(http.StatusBadRequest, "candidate_id is required")
	}

	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: candidateID},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}
	if user.SystemRole == nil || user.SystemRole.Name != "Candidate" {
		return response.Error(http.StatusBadRequest, "User is not a candidate")
	}

	offset := (page - 1) * limit

	total, err := u.gormDbRepo.CountOnboardHistoriesByCandidate(ctx, candidateID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count onboarding history")
	}

	rows, err := u.gormDbRepo.FetchOnboardHistoriesByCandidate(ctx, candidateID, limit, offset)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch onboarding history")
	}

	var results []interface{}
	for _, row := range rows {
		resp := row.ToOnboardHistoryResp()
		results = append(results, resp)
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

func (u *appUsecase) FetchOnboardingActiveTeam(ctx context.Context, employeeID string, page, limit int64, cursor string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if employeeID == "" {
		return response.Error(http.StatusBadRequest, "employee_id is required")
	}

	offset := (page - 1) * limit

	total, err := u.gormDbRepo.CountOnboardHistoriesByEmployee(ctx, employeeID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count onboarding history")
	}

	rows, err := u.gormDbRepo.FetchOnboardHistoriesByEmployee(ctx, employeeID, limit, offset)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch onboarding history")
	}

	var results []interface{}
	for _, row := range rows {
		resp := row.ToOnboardHistoryResp()
		results = append(results, resp)
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

func (u *appUsecase) FetchOnboardingHistory(ctx context.Context, employeeID string, page, limit int64, cursor string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if employeeID == "" {
		return response.Error(http.StatusBadRequest, "employee_id is required")
	}

	offset := (page - 1) * limit

	now := time.Now().UTC()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	total, err := u.gormDbRepo.CountOnboardHistoriesByEmployeeHistory(ctx, employeeID, currentDate)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count onboarding history")
	}

	rows, err := u.gormDbRepo.FetchOnboardHistoriesByEmployeeHistory(ctx, employeeID, currentDate, limit, offset)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch onboarding history")
	}

	var results []interface{}
	for _, row := range rows {
		resp := row.ToOnboardHistoryResp()
		results = append(results, resp)
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

func (u *appUsecase) FetchOnboardingActive(ctx context.Context, page, limit int64, cursor string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit
	now := time.Now().UTC()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	query := u.gormDbRepo.GetDB().WithContext(ctx).
		Model(&gorm_model.OnboardHistory{}).
		Where("start_date <= ? AND (end_date IS NULL OR end_date >= ?)", currentDate, currentDate)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count onboarding history")
	}

	var rows []gorm_model.OnboardHistory
	if err := query.
		Preload("CandidateUser").
		Preload("CandidateUser.SystemRole").
		Preload("Review").
		Order("start_date DESC NULLS LAST, created_at DESC").
		Limit(int(limit)).Offset(int(offset)).
		Find(&rows).Error; err != nil {
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch onboarding history: %v", err))
	}

	results := make([]interface{}, 0, len(rows))
	for _, row := range rows {
		resp := row.ToOnboardHistoryResp()
		results = append(results, resp)
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

func (u *appUsecase) FetchOnboardingArchive(ctx context.Context, page, limit int64, cursor string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit
	now := time.Now().UTC()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	query := u.gormDbRepo.GetDB().WithContext(ctx).
		Model(&gorm_model.OnboardHistory{}).
		Where("end_date < ?", currentDate)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count onboarding history")
	}

	var rows []gorm_model.OnboardHistory
	if err := query.
		Preload("CandidateUser").
		Preload("CandidateUser.SystemRole").
		Preload("Review").
		Order("end_date DESC NULLS LAST, created_at DESC").
		Limit(int(limit)).Offset(int(offset)).
		Find(&rows).Error; err != nil {
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch onboarding archive: %v", err))
	}

	results := make([]interface{}, 0, len(rows))
	for _, row := range rows {
		resp := row.ToOnboardHistoryResp()
		results = append(results, resp)
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
