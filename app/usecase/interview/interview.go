package usecase_interview

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func normalizeInterviewMethod(method string) (string, bool) {
	switch strings.ToLower(method) {
	case "online":
		return "Online", true
	case "offline":
		return "Offline", true
	default:
		return "", false
	}
}

func parseScheduledAt(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduled_at format, expected RFC3339")
	}
	return &t, nil
}

func (u *appUsecase) FetchAll(ctx context.Context, page, limit int64, cursor string, filter gorm_model.InterviewFilter) response.Base {
	return u.fetchList(ctx, page, limit, filter)
}

func (u *appUsecase) FetchScheduled(ctx context.Context, adminID string, page, limit int64, cursor string) response.Base {
	status := "SCHEDULED"
	filter := gorm_model.InterviewFilter{
		AdminUserID:   &adminID,
		ScheduledOnly: true,
		Status:        &status,
	}
	return u.fetchList(ctx, page, limit, filter)
}

func (u *appUsecase) fetchList(ctx context.Context, page, limit int64, filter gorm_model.InterviewFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	limitPtr := &limit
	offset := (page - 1) * limit
	filter.Limit = limitPtr
	filter.Offset = &offset
	if len(filter.Sorts) == 0 {
		filter.Sorts = []map[string]string{{"scheduled_at": "DESC"}, {"created_at": "DESC"}}
	}

	countFilter := filter
	countFilter.Limit = nil
	countFilter.Offset = nil

	dbCount := u.gormDbRepo.GetDB().WithContext(ctx).Model(&gorm_model.Interview{})
	countFilter.Query(dbCount)
	var total int64
	if err := dbCount.Count(&total).Error; err != nil {
		logrus.Error("Interview count error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to count Interview records")
	}

	var interviews []gorm_model.Interview
	dbFetch := u.gormDbRepo.GetDB().WithContext(ctx).
		Preload("Stage").
		Preload("Subrequest").
		Preload("Subrequest.Request").
		Preload("Subrequest.JobRole").
		Preload("CandidateUser")
	filter.Query(dbFetch)
	if err := dbFetch.Find(&interviews).Error; err != nil {
		logrus.Error("Interview fetch error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Interview data")
	}

	var results []interface{}
	for _, interview := range interviews {
		results = append(results, interview.ToInterviewResp())
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

	interview, err := u.gormDbRepo.GetInterviewByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Interview not found")
		}
		logrus.Error("Interview FetchData error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Interview")
	}

	return response.Success(interview.ToInterviewResp())
}

func (u *appUsecase) Create(ctx context.Context, adminID string, req request_model.CreateInterviewRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	scheduledAt, err := parseScheduledAt(req.ScheduledAt)
	if err != nil {
		return response.Error(http.StatusBadRequest, err.Error())
	}

	method, ok := normalizeInterviewMethod(req.Method)
	if !ok {
		return response.Error(http.StatusBadRequest, "Invalid method value. Allowed: Online, Offline")
	}

	if req.CandidateUserID == nil || *req.CandidateUserID == "" || req.SubrequestID == nil || *req.SubrequestID == "" {
		return response.Error(http.StatusBadRequest, "candidate_user_id and subrequest_id are required for new interviews")
	}

	if authRes := u.ensureAdminOfSubrequest(ctx, adminID, *req.SubrequestID); authRes.Status != http.StatusOK {
		return authRes
	}

	if candRes := u.ensureCandidateOnSubrequest(ctx, *req.CandidateUserID, *req.SubrequestID); candRes.Status != http.StatusOK {
		return candRes
	}

	var stage gorm_model.InterviewStage
	if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&stage, "id = ?", req.StageID).Error; err != nil {
		return response.Error(http.StatusBadRequest, "Invalid interview stage ID")
	}
	if !stage.IsActive {
		return response.Error(http.StatusBadRequest, "Cannot reference an inactive interview stage")
	}

	id := uuid.NewString()

	newInterview := gorm_model.Interview{
		ID:              id,
		AdminUserID:     &adminID,
		CandidateUserID: *req.CandidateUserID,
		SubrequestID:    *req.SubrequestID,
		StageID:         req.StageID,
		Title:           req.Title,
		Description:     req.Description,
		ScheduledAt:     scheduledAt,
		Method:          &method,
		Status:          "SCHEDULED",
		MeetingLink:     req.MeetingLink,
		MeetingLocation: req.MeetingLocation,
	}

	if err := u.gormDbRepo.CreateInterview(ctx, &newInterview); err != nil {
		logrus.Error("Interview Create error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Interview", id, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create Interview")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Create", "Interview", id, req, true)

	created, err := u.gormDbRepo.GetInterviewByID(ctx, id)
	if err == nil {
		return response.Success(created.ToInterviewResp())
	}

	return response.Success(newInterview.ToInterviewResp())
}

func (u *appUsecase) Update(ctx context.Context, adminID string, id string, req request_model.UpdateInterviewRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	scheduledAt, err := parseScheduledAt(req.ScheduledAt)
	if err != nil {
		return response.Error(http.StatusBadRequest, err.Error())
	}

	method, ok := normalizeInterviewMethod(req.Method)
	if !ok {
		return response.Error(http.StatusBadRequest, "Invalid method value. Allowed: Online, Offline")
	}

	existing, err := u.gormDbRepo.GetInterviewByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Interview not found")
		}
		logrus.Error("Interview Update fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Interview")
	}

	if authRes := u.ensureAdminOfSubrequest(ctx, adminID, existing.SubrequestID); authRes.Status != http.StatusOK {
		return authRes
	}

	if existing.StageID != req.StageID {
		var stage gorm_model.InterviewStage
		if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&stage, "id = ?", req.StageID).Error; err != nil {
			return response.Error(http.StatusBadRequest, "Invalid interview stage ID")
		}
		if !stage.IsActive {
			return response.Error(http.StatusBadRequest, "Cannot reference an inactive interview stage")
		}
	}

	existing.StageID = req.StageID
	existing.Title = req.Title
	existing.Description = req.Description
	existing.ScheduledAt = scheduledAt
	existing.Method = &method
	existing.MeetingLocation = req.MeetingLocation
	existing.MeetingLink = req.MeetingLink
	existing.AdminUserID = &adminID

	if err := u.gormDbRepo.UpdateInterview(ctx, existing); err != nil {
		logrus.Error("Interview Update error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Interview", id, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to update Interview")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Interview", id, req, true)

	updated, err := u.gormDbRepo.GetInterviewByID(ctx, id)
	if err == nil {
		return response.Success(updated.ToInterviewResp())
	}

	return response.Success(existing.ToInterviewResp())
}

func (u *appUsecase) PatchStage(ctx context.Context, adminID string, req request_model.PatchInterviewStageRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	interview, err := u.gormDbRepo.GetInterviewByID(ctx, req.InterviewID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Interview not found")
		}
		logrus.Error("Interview PatchStage fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Interview")
	}

	if authRes := u.ensureAdminOfSubrequest(ctx, adminID, interview.SubrequestID); authRes.Status != http.StatusOK {
		return authRes
	}

	if interview.StageID != req.StageID {
		var stage gorm_model.InterviewStage
		if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&stage, "id = ?", req.StageID).Error; err != nil {
			return response.Error(http.StatusBadRequest, "Invalid interview stage ID")
		}
		if !stage.IsActive {
			return response.Error(http.StatusBadRequest, "Cannot reference an inactive interview stage")
		}
	}

	if err := u.gormDbRepo.PatchInterviewStage(ctx, req.InterviewID, req.StageID, req.Status); err != nil {
		logrus.Error("Interview PatchStage error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Update Stage", "Interview", req.InterviewID, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to update Interview stage")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update Stage", "Interview", req.InterviewID, req, true)

	updated, err := u.gormDbRepo.GetInterviewByID(ctx, req.InterviewID)
	if err == nil {
		return response.Success(updated.ToInterviewResp())
	}

	return response.Success(nil)
}

func (u *appUsecase) PatchStatus(ctx context.Context, adminID string, req request_model.PatchInterviewStatusRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	interview, err := u.gormDbRepo.GetInterviewByID(ctx, req.InterviewID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Interview not found")
		}
		logrus.Error("Interview PatchStatus fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch Interview")
	}

	if authRes := u.ensureAdminOfSubrequest(ctx, adminID, interview.SubrequestID); authRes.Status != http.StatusOK {
		return authRes
	}

	if err := u.gormDbRepo.PatchInterviewStatus(ctx, req.InterviewID, req.Status); err != nil {
		logrus.Error("Interview PatchStatus error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Update Status", "Interview", req.InterviewID, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to update Interview status")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update Status", "Interview", req.InterviewID, req, true)

	updated, err := u.gormDbRepo.GetInterviewByID(ctx, req.InterviewID)
	if err == nil {
		return response.Success(updated.ToInterviewResp())
	}

	return response.Success(nil)
}

func (u *appUsecase) ensureAdminOfSubrequest(ctx context.Context, adminID string, subrequestID string) response.Base {
	isAdmin, err := u.gormDbRepo.IsAdminOfSubrequest(ctx, adminID, subrequestID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify admin assignment")
	}
	if !isAdmin {
		return response.Error(http.StatusForbidden, "You are not assigned to this subrequest's request")
	}
	return response.Success(nil)
}

func (u *appUsecase) ensureCandidateOnSubrequest(ctx context.Context, candidateID string, subrequestID string) response.Base {
	isCandidate, err := u.gormDbRepo.IsCandidateOnSubrequest(ctx, candidateID, subrequestID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to verify candidate assignment")
	}
	if !isCandidate {
		return response.Error(http.StatusBadRequest, "Candidate is not assigned to this subrequest")
	}
	return response.Success(nil)
}