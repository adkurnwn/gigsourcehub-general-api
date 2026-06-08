package usecase_request

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
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

		var jr gorm_model.JobRole
		if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&jr, "id = ?", sub.JobRoleID).Error; err != nil {
			return response.Error(http.StatusBadRequest, "Invalid job role ID")
		}
		if !jr.IsActive {
			return response.Error(http.StatusBadRequest, "Cannot reference an inactive job role")
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

	// 5. Notify all admin users about the new request
	_, employee, _ := u.verifyEmployee(ctx, employeeID)
	employeeName := employeeID
	if employee != nil {
		employeeName = employee.Name
	}
	go func() {
		bgCtx := context.Background()
		var adminUserIDs []string
		err := u.gormDbRepo.GetDB().WithContext(bgCtx).
			Table("users").
			Joins("JOIN system_roles sr ON sr.id = users.system_role_id").
			Where("sr.name = ? AND users.deleted_at IS NULL", "Admin").
			Pluck("users.id", &adminUserIDs).Error
		
		if err == nil && len(adminUserIDs) > 0 {
			title := "New Request Submitted"
			desc := fmt.Sprintf("A new request has been submitted by %s.", employeeName)
			helpers.SendNotificationToAll(bgCtx, u.gormDbRepo, adminUserIDs, title, desc)
		}
	}()

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

// verifyAdmin checks if the user has the "Admin" system role
func (u *appUsecase) verifyAdmin(ctx context.Context, userID string) (bool, *gorm_model.User, error) {
	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: userID},
	})
	if err != nil {
		return false, nil, err
	}
	if user == nil {
		return false, nil, nil
	}

	if user.SystemRole != nil && user.SystemRole.Name == "Admin" {
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

func (u *appUsecase) FetchAll(ctx context.Context, page, limit int64) response.Base {
	return u.fetchByAdminWithFilter(ctx, page, limit, gorm_model.RequestFilter{})
}

func (u *appUsecase) FetchByAdmin(ctx context.Context, page, limit int64, filter gorm_model.RequestFilter) response.Base {
	return u.fetchByAdminWithFilter(ctx, page, limit, filter)
}

func (u *appUsecase) FetchPendingForAdmin(ctx context.Context, page, limit int64, filter gorm_model.RequestFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit
	status := "PENDING"
	filter.Status = &status

	total, err := u.gormDbRepo.CountRequestsByAdmin(ctx, filter)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count requests")
	}

	rows, err := u.gormDbRepo.FetchRequestsByAdmin(ctx, filter, limit, offset)
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

func (u *appUsecase) FetchMyRequestsForAdmin(ctx context.Context, adminID string, page, limit int64, filter gorm_model.RequestFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit
	filter.AdminUserID = &adminID

	total, err := u.gormDbRepo.CountRequestsByAdmin(ctx, filter)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count requests")
	}

	rows, err := u.gormDbRepo.FetchRequestsByAdmin(ctx, filter, limit, offset)
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

func (u *appUsecase) fetchByAdminWithFilter(ctx context.Context, page, limit int64, filter gorm_model.RequestFilter) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit

	total, err := u.gormDbRepo.CountRequestsByAdmin(ctx, filter)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to count requests")
	}

	rows, err := u.gormDbRepo.FetchRequestsByAdmin(ctx, filter, limit, offset)
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

	isAdmin, _, err := u.verifyAdmin(ctx, employeeID)
	if err != nil {
		logrus.Errorf("Failed to verify admin role: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to verify user")
	}

	// Make sure the employee who made the request is the one retrieving it, unless admin
	if !isAdmin && req.EmployeeUserID != employeeID {
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

func (u *appUsecase) AssignPIC(ctx context.Context, adminID, requestID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	existingReq, err := u.gormDbRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		if err.Error() == "record not found" {
			return response.Error(http.StatusNotFound, "Request not found")
		}
		return response.Error(http.StatusInternalServerError, "Failed to fetch request")
	}

	if existingReq.AdminUserID != nil && *existingReq.AdminUserID != adminID {
		return response.Error(http.StatusConflict, "Request already assigned to another admin")
	}

	// Set admin user and change status to ACCEPTED when assigning PIC
	existingReq.AdminUserID = &adminID
	existingReq.Status = "ACCEPTED"

	if err := u.gormDbRepo.UpdateRequestByAdmin(ctx, existingReq); err != nil {
		logrus.Errorf("AssignPIC DB Error: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to assign PIC")
	}

	// Notify the employee that their request was accepted
	helpers.SendNotificationAsync(ctx, u.gormDbRepo, existingReq.EmployeeUserID,
		"Request Accepted",
		fmt.Sprintf("Your request \"%s\" has been accepted by an admin.", existingReq.ProjectName),
	)

	return response.Success(nil)
}

func (u *appUsecase) RejectRequest(ctx context.Context, adminID, requestID string, rejectedReason string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	existingReq, err := u.gormDbRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		if err.Error() == "record not found" {
			return response.Error(http.StatusNotFound, "Request not found")
		}
		return response.Error(http.StatusInternalServerError, "Failed to fetch request")
	}

	if existingReq.Status != "PENDING" {
		return response.Error(http.StatusConflict, "Only PENDING requests can be rejected")
	}

	if rejectedReason == "" {
		return response.Error(http.StatusBadRequest, "rejected_reason is required")
	}

	existingReq.Status = "REJECTED"
	existingReq.RejectedReason = &rejectedReason

	if err := u.gormDbRepo.UpdateRequestByAdmin(ctx, existingReq); err != nil {
		logrus.Errorf("RejectRequest DB Error: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to reject request")
	}

	// Notify the employee that their request was rejected
	helpers.SendNotificationAsync(ctx, u.gormDbRepo, existingReq.EmployeeUserID,
		"Request Rejected",
		fmt.Sprintf("Your request \"%s\" has been rejected. Reason: %s", existingReq.ProjectName, rejectedReason),
	)

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

	if req.JobRoleID != nil && *req.JobRoleID != "" {
		if existingSubReq.JobRoleID == nil || *existingSubReq.JobRoleID != *req.JobRoleID {
			var jr gorm_model.JobRole
			if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&jr, "id = ?", *req.JobRoleID).Error; err != nil {
				return response.Error(http.StatusBadRequest, "Invalid job role ID")
			}
			if !jr.IsActive {
				return response.Error(http.StatusBadRequest, "Cannot reference an inactive job role")
			}
		}
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

	var jr gorm_model.JobRole
	if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&jr, "id = ?", req.JobRoleID).Error; err != nil {
		return response.Error(http.StatusBadRequest, "Invalid job role ID")
	}
	if !jr.IsActive {
		return response.Error(http.StatusBadRequest, "Cannot reference an inactive job role")
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

func (u *appUsecase) AssignCandidateToSubrequest(ctx context.Context, adminID string, requestID string, subrequestID string, req request_model.AssignCandidateToSubrequestRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if req.CandidateUserID == "" {
		return response.Error(http.StatusBadRequest, "candidate_user_id is required")
	}

	existingReq, err := u.gormDbRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		if err.Error() == "record not found" {
			return response.Error(http.StatusNotFound, "Request not found")
		}
		return response.Error(http.StatusInternalServerError, "Failed to fetch parent request")
	}

	if existingReq.AdminUserID == nil || *existingReq.AdminUserID != adminID {
		return response.Error(http.StatusForbidden, "You do not have permission to assign candidates to this request")
	}

	existingSubReq, err := u.gormDbRepo.GetSubrequestByID(ctx, subrequestID)
	if err != nil {
		if err.Error() == "record not found" {
			return response.Error(http.StatusNotFound, "Subrequest not found")
		}
		return response.Error(http.StatusInternalServerError, "Failed to fetch subrequest")
	}

	if existingSubReq.RequestID != requestID {
		return response.Error(http.StatusBadRequest, "Subrequest does not belong to the targeted Request ID")
	}

	candidate, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: req.CandidateUserID},
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to fetch candidate")
	}
	if candidate == nil {
		return response.Error(http.StatusNotFound, "Candidate not found")
	}
	if candidate.SystemRole == nil || candidate.SystemRole.Name != "Candidate" {
		return response.Error(http.StatusBadRequest, "User is not a candidate")
	}

	statusName, err := u.gormDbRepo.GetCandidateRecruitmentStatusName(ctx, candidate.ID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to check recruitment status")
	}

	activeAssignments, err := u.gormDbRepo.CountActiveSubrequestCandidatesByCandidateID(ctx, candidate.ID)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to check candidate assignments")
	}
	if statusName != "" {
		return response.Error(http.StatusConflict, "Candidate recruitment status must be null to assign")
	}
	if activeAssignments > 0 {
		return response.Error(http.StatusConflict, "Candidate is still in a recruitment process")
	}

	var assignedStatus gorm_model.RecruitmentStatus
	if err := u.gormDbRepo.GetDB().WithContext(ctx).Where("name = ?", "Assigned").First(&assignedStatus).Error; err != nil {
		return response.Error(http.StatusInternalServerError, "Assigned recruitment status not found")
	}

	assignment := &gorm_model.SubrequestCandidate{
		SubrequestID:    subrequestID,
		CandidateUserID: candidate.ID,
		Name:            candidate.Name,
	}
	markRequestProcessing := existingReq.Status != "PROCESSING"

	if err := u.gormDbRepo.AssignCandidateToSubrequest(ctx, assignment, assignedStatus.ID, requestID, markRequestProcessing); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already assigned") {
			return response.Error(http.StatusConflict, "Candidate is already assigned to this subrequest")
		}
		logrus.Errorf("AssignCandidateToSubrequest DB Error: %v", err)
		return response.Error(http.StatusInternalServerError, "Failed to assign candidate to subrequest")
	}

	// Notify the candidate about being recruited for a position
	var jobRoleName string
	if existingSubReq.JobRoleID != nil {
		var jr gorm_model.JobRole
		if err := u.gormDbRepo.GetDB().WithContext(ctx).First(&jr, "id = ?", *existingSubReq.JobRoleID).Error; err == nil {
			jobRoleName = jr.Name
		}
	}
	helpers.SendNotificationAsync(ctx, u.gormDbRepo, candidate.ID,
		"Recruitment Invitation",
		fmt.Sprintf("You have been invited to the recruitment process for the position of %s. Please check your dashboard for further information.", jobRoleName),
	)

	// Send Email asynchronously
	go func(email, name, roleName string) {
		if err := u.mailerRepo.SendRecruitmentInvitationEmail(email, name, roleName); err != nil {
			logrus.Errorf("Failed to send recruitment invitation email to %s: %v", email, err)
		}
	}(candidate.Email, candidate.Name, jobRoleName)

	return response.Success(nil)
}
