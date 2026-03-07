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
		DueDate:           dueDate,
		EmployeeUserID:    employeeID,
		RequiredHeadcount: len(req.Subrequests),
		Status:            "Pending", // Default status for new requests
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
			MinYearsExperience: sub.MinYearsExperience,
			JobTitleID:         sub.JobTitleID,
			TechStack:          techStackJSON,
			Notes:              sub.Notes,
			IsFilled:           false, // Default value
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
