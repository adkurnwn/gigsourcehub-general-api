package usecase_company_profile

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ---------- Admin & Superadmin — CMS ----------

// Get returns the current company profile data.
func (u *appUsecase) Get(ctx context.Context) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	profile, err := u.gormDbRepo.GetCompanyProfile(ctx)
	if err != nil {
		logrus.Error("CompanyProfile Get error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch company profile")
	}

	return response.Success(profile.ToCompanyProfileResp())
}

// ---------- Admin only ----------

// AdminUpdate creates an approval request for updating the company profile.
// Does NOT update the company_profiles table directly.
// If there's already a PENDING request, admin must cancel it first.
func (u *appUsecase) AdminUpdate(ctx context.Context, adminID string, req request_model.UpdateCompanyProfileRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Fetch current profile to get its ID
	profile, err := u.gormDbRepo.GetCompanyProfile(ctx)
	if err != nil {
		logrus.Error("CompanyProfile AdminUpdate fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch company profile")
	}

	// Check if there's already a PENDING approval for this record
	existing, err := u.gormDbRepo.GetPendingApprovalByRecord(ctx, "company_profiles", profile.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Error("CompanyProfile AdminUpdate pending check error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to check pending approvals")
	}
	if existing != nil {
		return response.Error(http.StatusConflict, "There is already a pending update for company profile. Please cancel it before creating a new one.")
	}

	// Serialize proposed data
	proposedJSON, err := json.Marshal(req)
	if err != nil {
		logrus.Error("CompanyProfile AdminUpdate marshal error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to process company profile data")
	}
	proposedStr := string(proposedJSON)

	approval := &gorm_model.ApprovalRequest{
		RequestedByAdminID: adminID,
		TableName:          "company_profiles",
		RecordID:           profile.ID,
		Action:             "UPDATE",
		ProposedData:       &proposedStr,
		Status:             "PENDING",
	}

	if err := u.gormDbRepo.CreateApprovalRequest(ctx, approval); err != nil {
		logrus.Error("CompanyProfile AdminUpdate approval error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Company Profile", "", req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create company profile update approval request")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update", "Company Profile", "", req, true)
	return response.Success(approval.ToApprovalRequestResp())
}

// CancelPendingApproval allows Admin to cancel a pending company profile approval.
func (u *appUsecase) CancelPendingApproval(ctx context.Context, approvalID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	approval, err := u.gormDbRepo.GetApprovalRequestByID(ctx, approvalID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Approval request not found")
		}
		logrus.Error("CompanyProfile CancelPending fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch approval request")
	}

	if approval.TableName != "company_profiles" {
		return response.Error(http.StatusBadRequest, "This approval request is not for company profile")
	}

	if approval.Status != "PENDING" {
		return response.Error(http.StatusBadRequest, "Only PENDING approval requests can be cancelled")
	}

	if err := u.gormDbRepo.DeleteApprovalRequest(ctx, approvalID); err != nil {
		logrus.Error("CompanyProfile CancelPending delete error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to cancel approval request")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Cancel", "Company Profile Approval", approvalID, nil, true)
	return response.Success(nil)
}

// ---------- Superadmin — Approvals ----------

// FetchPendingApprovals returns all approval requests for company_profiles.
func (u *appUsecase) FetchPendingApprovals(ctx context.Context, page, limit int64) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit
	tableName := "company_profiles"
	filter := gorm_model.ApprovalRequestFilter{
		TableNameEq: &tableName,
	}
	filter.Limit = &limit
	filter.Offset = &offset
	if len(filter.Sorts) == 0 {
		filter.Sorts = []map[string]string{
			{"created_at": "DESC"},
		}
	}

	total, err := u.gormDbRepo.CountApprovalRequests(ctx, gorm_model.ApprovalRequestFilter{TableNameEq: &tableName})
	if err != nil {
		logrus.Error("CompanyProfile FetchPendingApprovals count error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to count approval requests")
	}

	requests, err := u.gormDbRepo.FetchApprovalRequests(ctx, filter)
	if err != nil {
		logrus.Error("CompanyProfile FetchPendingApprovals fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch approval requests")
	}

	var results []interface{}
	for _, r := range requests {
		results = append(results, r.ToApprovalRequestResp())
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

// ApproveRequest — Superadmin approves a company profile update.
func (u *appUsecase) ApproveRequest(ctx context.Context, superadminID string, approvalID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	approval, err := u.gormDbRepo.GetApprovalRequestByID(ctx, approvalID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Approval request not found")
		}
		logrus.Error("CompanyProfile ApproveRequest fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch approval request")
	}

	if approval.TableName != "company_profiles" {
		return response.Error(http.StatusBadRequest, "This approval request is not for company profile")
	}

	if approval.Status != "PENDING" {
		return response.Error(http.StatusBadRequest, "Approval request is not in PENDING status")
	}

	// Parse proposed data
	var proposedData request_model.UpdateCompanyProfileRequest
	if approval.ProposedData != nil {
		if err := json.Unmarshal([]byte(*approval.ProposedData), &proposedData); err != nil {
			logrus.Error("CompanyProfile ApproveRequest unmarshal error:", err)
			return response.Error(http.StatusInternalServerError, "Failed to parse proposed data")
		}
	}

	// Fetch current profile
	profile, err := u.gormDbRepo.GetCompanyProfile(ctx)
	if err != nil {
		logrus.Error("CompanyProfile ApproveRequest profile fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch company profile")
	}

	// Apply proposed changes
	profile.Address = proposedData.Address
	profile.Phone = proposedData.Phone
	profile.Email = proposedData.Email
	profile.FacebookURL = proposedData.FacebookURL
	profile.InstagramURL = proposedData.InstagramURL
	profile.LinkedinURL = proposedData.LinkedinURL
	profile.TwitterURL = proposedData.TwitterURL

	if err := u.gormDbRepo.UpdateCompanyProfile(ctx, profile); err != nil {
		logrus.Error("CompanyProfile ApproveRequest update error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to update company profile")
	}

	// Mark approval as APPROVED
	approval.Status = "APPROVED"
	approval.ReviewedBySuperadminID = &superadminID
	if err := u.gormDbRepo.UpdateApprovalRequest(ctx, approval); err != nil {
		logrus.Error("CompanyProfile ApproveRequest status update error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to update approval status")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Approve", "Company Profile", "", nil, true)
	return response.Success(approval.ToApprovalRequestResp())
}

// RejectRequest — Superadmin rejects a company profile update.
func (u *appUsecase) RejectRequest(ctx context.Context, superadminID string, approvalID string, req request_model.ReviewApprovalRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	approval, err := u.gormDbRepo.GetApprovalRequestByID(ctx, approvalID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Approval request not found")
		}
		logrus.Error("CompanyProfile RejectRequest fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch approval request")
	}

	if approval.TableName != "company_profiles" {
		return response.Error(http.StatusBadRequest, "This approval request is not for company profile")
	}

	if approval.Status != "PENDING" {
		return response.Error(http.StatusBadRequest, "Approval request is not in PENDING status")
	}

	approval.Status = "REJECTED"
	approval.ReviewedBySuperadminID = &superadminID
	approval.RejectedReason = req.RejectedReason

	if err := u.gormDbRepo.UpdateApprovalRequest(ctx, approval); err != nil {
		logrus.Error("CompanyProfile RejectRequest update error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to update approval status")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Reject", "Company Profile", "", req, true)
	return response.Success(approval.ToApprovalRequestResp())
}

// ---------- Public — no auth ----------

// GetPublic returns the current company profile for public consumption.
func (u *appUsecase) GetPublic(ctx context.Context) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	profile, err := u.gormDbRepo.GetCompanyProfile(ctx)
	if err != nil {
		logrus.Error("CompanyProfile GetPublic error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch company profile")
	}

	return response.Success(profile.ToCompanyProfileResp())
}
