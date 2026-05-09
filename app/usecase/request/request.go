package usecase_request

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/sirupsen/logrus"
)

func (u *appUsecase) CreateByEmployee(ctx context.Context, employeeID string, req request_model.CreateRequestRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1. Verify that the user exists and is actually an employee
	isEmployee, _, err := u.verifyEmployee(ctx, employeeID)
	if err != nil {
		logrus.Errorf("Failed to verify employee role: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to verify user")
	}
	if !isEmployee {
		return response.Error(http.StatusForbidden, "Only employees can create requests")
	}

	// 2. Map payload to GORM models
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *req.DueDate)
		if err == nil {
			dueDate = &parsedDate
		} else {
			return response.Error(http.StatusBadRequest, "Invalid due_date format, expected YYYY-MM-DD")
		}
	}

	requestModel := &gorm_model.Request{
		ProjectName:       req.ProjectName,
		ProjectDuration:   req.ProjectDuration,
		DueDate:           dueDate,
		EmployeeUserID:    employeeID,
		RequiredHeadcount: len(req.Subrequests),
		Status:            "PENDING", // Default status for new requests
		Urgency:           req.Urgency,
	}

	// 3. Map subrequests
	var subrequestModels []gorm_model.Subrequest
	for _, sub := range req.Subrequests {
		var techStackJSON *string
		if len(sub.TechStack) > 0 {
			b, err := json.Marshal(sub.TechStack)
			if err == nil {
				jsonStr := string(b)
				techStackJSON = &jsonStr
			}
		}

		subReq := gorm_model.Subrequest{
			Level:     &sub.Level,
			JobRoleID: sub.JobRoleID,
			TechStack: techStackJSON,
			Notes:     sub.Notes,
			Overview:  sub.Overview,
			IsFilled:  false, // Default value
		}
		subrequestModels = append(subrequestModels, subReq)
	}
	requestModel.Subrequests = subrequestModels

	// 4. Save to Database (GORM automatically inserts the parent Request and the child Subrequests in a transaction)
	if err := u.gormDbRepo.CreateRequest(ctx, requestModel); err != nil {
		logrus.Errorf("Failed to create request: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to create request")
	}

	return response.Success(nil)
}

// verifyEmployee checks if the user has the "Employee" system role
func (u *appUsecase) verifyEmployee(ctx context.Context, userID string) (bool, *gorm_model.User, error) {
	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: userID},
	})
	if err != nil {
		return false, nil, err
	}
	if user == nil {
		return false, nil, nil
	}

	if user.SystemRole != nil && user.SystemRole.Name == "Employee" {
		return true, user, nil
	}
	return false, user, nil
}

func (u *appUsecase) FetchByEmployee(ctx context.Context, employeeID string, page, limit int64) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit

	total, err := u.gormDbRepo.CountRequestsByEmployee(ctx, employeeID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count requests")
	}

	rows, err := u.gormDbRepo.FetchRequestsByEmployee(ctx, employeeID, limit, offset)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch requests")
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var req gorm_model.Request
		if err := u.gormDbRepo.StructScan(rows, &req); err != nil {
			logrus.Errorf("Failed to scan request: %v", err)
			continue
		}

		// Because StructScan doesn't execute Preloads, we need to fetch the whole struct natively
		// Preloads only work on DB.Find() and First(), not strictly raw Row iterating unless handled with Gorm directly.
		// A cleaner standard GORM pagination approach would use Find... Let's just lookup by ID real quick for each since they're paginated to limit 10:
		fullReq, err := u.gormDbRepo.GetRequestByID(ctx, req.ID)
		if err == nil {
			results = append(results, fullReq.ToRequestResp())
		}
	}

	return response.Success(response.List{
		List:  results,
		Limit: limit,
		Page:  page,
		Total: total,
	})
}

func (u *appUsecase) GetByID(ctx context.Context, employeeID, requestID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	req, err := u.gormDbRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		// Differentiate between generic DB error and Not Found
		if err.Error() == "record not found" {
			return response.Error(http.StatusNotFound, "Request not found")
		}
		return response.Error(http.StatusInternalServerError, "Failed to fetch request")
	}

	// Make sure the employee who made the request is the one retrieving it
	if req.EmployeeUserID != employeeID {
		return response.Error(http.StatusForbidden, "You do not have permission to view this request")
	}

	return response.Success(req.ToRequestResp())
}

func (u *appUsecase) UpdateByEmployee(ctx context.Context, employeeID string, requestID string, req request_model.UpdateRequestRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1. Fetch Existing Request
	existingReq, err := u.gormDbRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		if err.Error() == "record not found" {
			return response.Error(http.StatusNotFound, "Request not found")
		}
		return response.Error(http.StatusInternalServerError, "Failed to fetch request")
	}

	// 2. Validate Ownership
	if existingReq.EmployeeUserID != employeeID {
		return response.Error(http.StatusForbidden, "You do not have permission to edit this request")
	}

	// 3. Validate Status
	if existingReq.Status != "PENDING" && existingReq.Status != "WAITING" {
		return response.Error(http.StatusConflict, "Only PENDING requests can be edited")
	}

	// 4. Map updated main fields
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *req.DueDate)
		if err == nil {
			dueDate = &parsedDate
		} else {
			return response.Error(http.StatusBadRequest, "Invalid due_date format, expected YYYY-MM-DD")
		}
	} else {
		dueDate = existingReq.DueDate
	}

	existingReq.ProjectName = req.ProjectName
	existingReq.ProjectDuration = req.ProjectDuration
	existingReq.Urgency = req.Urgency
	existingReq.DueDate = dueDate

	// Execute update
	if err := u.gormDbRepo.UpdateRequestByEmployee(ctx, existingReq); err != nil {
		logrus.Errorf("UpdateByEmployee DB Error: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to update request")
	}

	return response.Success(nil)
}

func (u *appUsecase) UpdateSubrequestByEmployee(ctx context.Context, employeeID string, requestID string, subrequestID string, req request_model.UpdateSubrequestRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1. Fetch Existing Request for Authorization
	existingReq, err := u.gormDbRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		if err.Error() == "record not found" {
			return response.Error(http.StatusNotFound, "Request not found")
		}
		return response.Error(http.StatusInternalServerError, "Failed to fetch parent request")
	}

	// 2. Validate Ownership of the Parent Request
	if existingReq.EmployeeUserID != employeeID {
		return response.Error(http.StatusForbidden, "You do not have permission to edit subrequests belonging to this request")
	}

	// 3. Validate Status
	if existingReq.Status != "PENDING" && existingReq.Status != "WAITING" {
		return response.Error(http.StatusConflict, "Subrequests can only be edited when the parent request is PENDING")
	}

	// 4. Fetch target Subrequest
	existingSubReq, err := u.gormDbRepo.GetSubrequestByID(ctx, subrequestID)
	if err != nil {
		if err.Error() == "record not found" {
			return response.Error(http.StatusNotFound, "Subrequest not found")
		}
		return response.Error(http.StatusInternalServerError, "Failed to fetch subrequest")
	}

	// 5. Hard verify relation (Subrequest truly belongs to the requested Parent ID)
	if existingSubReq.RequestID != requestID {
		return response.Error(http.StatusBadRequest, "Subrequest does not belong to the targeted Request ID")
	}

	// 6. Map updated fields
	var techStackJSON *string
	if len(req.TechStack) > 0 {
		b, err := json.Marshal(req.TechStack)
		if err == nil {
			jsonStr := string(b)
			techStackJSON = &jsonStr
		}
	}

	existingSubReq.Level = &req.Level
	existingSubReq.JobRoleID = req.JobRoleID
	existingSubReq.TechStack = techStackJSON
	existingSubReq.Notes = req.Notes
	existingSubReq.Overview = req.Overview

	// Execute update specific to this Subrequest
	if err := u.gormDbRepo.UpdateSubrequestByEmployee(ctx, existingSubReq); err != nil {
		logrus.Errorf("UpdateSubrequestByEmployee DB Error: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to update subrequest")
	}

	return response.Success(nil)
}

func (u *appUsecase) AddSubrequestByEmployee(ctx context.Context, employeeID string, requestID string, req request_model.CreateSubrequestRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1. Fetch Existing Parent Request
	existingReq, err := u.gormDbRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		if err.Error() == "record not found" {
			return response.Error(http.StatusNotFound, "Request not found")
		}
		return response.Error(http.StatusInternalServerError, "Failed to fetch parent request")
	}

	// 2. Validate Ownership of the Parent Request
	if existingReq.EmployeeUserID != employeeID {
		return response.Error(http.StatusForbidden, "You do not have permission to add subrequests to this request")
	}

	// 3. Validate Status
	if existingReq.Status != "PENDING" && existingReq.Status != "WAITING" {
		return response.Error(http.StatusConflict, "Subrequests can only be added to PENDING requests")
	}

	// 4. Map Payload
	var techStackJSON *string
	if len(req.TechStack) > 0 {
		b, err := json.Marshal(req.TechStack)
		if err == nil {
			jsonStr := string(b)
			techStackJSON = &jsonStr
		}
	}

	subReq := &gorm_model.Subrequest{
		RequestID: existingReq.ID,
		Level:     &req.Level,
		JobRoleID: req.JobRoleID,
		TechStack: techStackJSON,
		Notes:     req.Notes,
		Overview:  req.Overview,
		IsFilled:  false,
	}

	// 5. Execute DB Transaction (Insert subrequest + Update master headcount)
	if err := u.gormDbRepo.CreateSubrequestByEmployee(ctx, subReq); err != nil {
		logrus.Errorf("AddSubrequestByEmployee DB Error: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to append subrequest")
	}

	return response.Success(nil)
}
