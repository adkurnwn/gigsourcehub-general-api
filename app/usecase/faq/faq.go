package usecase_faq

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ---------- Admin & Superadmin — CMS ----------

// FetchAll returns all FAQ records from the DB (these are all approved).
func (u *appUsecase) FetchAll(ctx context.Context, page, limit int64, search *string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit
	filter := gorm_model.FAQFilter{
		Search: search,
	}
	filter.Limit = &limit
	filter.Offset = &offset
	if len(filter.Sorts) == 0 {
		filter.Sorts = []map[string]string{
			{"created_at": "DESC"},
		}
	}

	// Count total
	total, err := u.gormDbRepo.CountFAQ(ctx, gorm_model.FAQFilter{Search: search})
	if err != nil {
		logrus.Error("FAQ FetchAll count error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to count FAQ records")
	}

	// Fetch paginated
	faqs, err := u.gormDbRepo.FetchFAQ(ctx, filter)
	if err != nil {
		logrus.Error("FAQ FetchAll fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch FAQ data")
	}

	var results []interface{}
	for _, f := range faqs {
		results = append(results, f.ToFAQResp())
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

// FetchData returns a single FAQ by ID.
func (u *appUsecase) FetchData(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	faq, err := u.gormDbRepo.GetFAQByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "FAQ not found")
		}
		logrus.Error("FAQ FetchData error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch FAQ")
	}

	return response.Success(faq.ToFAQResp())
}

// ---------- Admin only ----------

// Create — Admin creates a FAQ approval request (CREATE action).
// The actual FAQ is NOT inserted into the faqs table yet.
func (u *appUsecase) Create(ctx context.Context, adminID string, req request_model.CreateFAQRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Serialize proposed data
	proposedJSON, err := json.Marshal(map[string]string{
		"question": req.Question,
		"answer":   req.Answer,
	})
	if err != nil {
		logrus.Error("FAQ Create marshal error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to process FAQ data")
	}
	proposedStr := string(proposedJSON)

	// Generate a future record ID for the FAQ
	recordID := uuid.New().String()

	approval := &gorm_model.ApprovalRequest{
		RequestedByAdminID: adminID,
		TableName:          "faqs",
		RecordID:           recordID,
		Action:             "CREATE",
		ProposedData:       &proposedStr,
		Status:             "PENDING",
	}

	if err := u.gormDbRepo.CreateApprovalRequest(ctx, approval); err != nil {
		logrus.Error("FAQ Create approval error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Create", "FAQ", req.Question, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create FAQ approval request")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Create", "FAQ", req.Question, req, true)
	return response.Success(approval.ToApprovalRequestResp())
}

// Update — Admin creates a FAQ approval request (UPDATE action).
func (u *appUsecase) Update(ctx context.Context, adminID string, id string, req request_model.UpdateFAQRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Check if FAQ exists
	existing, err := u.gormDbRepo.GetFAQByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "FAQ not found")
		}
		logrus.Error("FAQ Update fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch FAQ")
	}

	// Serialize proposed data
	proposedJSON, err := json.Marshal(map[string]string{
		"question": req.Question,
		"answer":   req.Answer,
	})
	if err != nil {
		logrus.Error("FAQ Update marshal error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to process FAQ data")
	}
	proposedStr := string(proposedJSON)

	approval := &gorm_model.ApprovalRequest{
		RequestedByAdminID: adminID,
		TableName:          "faqs",
		RecordID:           existing.ID,
		Action:             "UPDATE",
		ProposedData:       &proposedStr,
		Status:             "PENDING",
	}

	if err := u.gormDbRepo.CreateApprovalRequest(ctx, approval); err != nil {
		logrus.Error("FAQ Update approval error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Update", "FAQ", existing.Question, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create FAQ update approval request")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update", "FAQ", existing.Question, req, true)
	return response.Success(approval.ToApprovalRequestResp())
}

// Delete — Admin directly soft-deletes a FAQ (no approval needed).
func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	existing, err := u.gormDbRepo.GetFAQByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "FAQ not found")
		}
		logrus.Error("FAQ Delete fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch FAQ")
	}

	if err := u.gormDbRepo.DeleteFAQ(ctx, id); err != nil {
		logrus.Error("FAQ Delete error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "FAQ", existing.Question, nil, false)
		return response.Error(http.StatusInternalServerError, "Failed to delete FAQ")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "FAQ", existing.Question, nil, true)
	return response.Success(nil)
}

// ---------- Superadmin — Approvals ----------

// FetchApprovals returns all approval requests for FAQs.
func (u *appUsecase) FetchApprovals(ctx context.Context, page, limit int64) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit
	tableName := "faqs"
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
		logrus.Error("FAQ FetchApprovals count error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to count approval requests")
	}

	requests, err := u.gormDbRepo.FetchApprovalRequests(ctx, filter)
	if err != nil {
		logrus.Error("FAQ FetchApprovals fetch error:", err)
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

// ApproveRequest — Superadmin approves a FAQ approval request.
func (u *appUsecase) ApproveRequest(ctx context.Context, superadminID string, approvalID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	approval, err := u.gormDbRepo.GetApprovalRequestByID(ctx, approvalID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Approval request not found")
		}
		logrus.Error("FAQ ApproveRequest fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch approval request")
	}

	if approval.TableName != "faqs" {
		return response.Error(http.StatusBadRequest, "This approval request is not for FAQs")
	}

	if approval.Status != "PENDING" {
		return response.Error(http.StatusBadRequest, "Approval request is not in PENDING status")
	}

	// Apply the action
	switch approval.Action {
	case "CREATE":
		var data map[string]string
		if approval.ProposedData != nil {
			if err := json.Unmarshal([]byte(*approval.ProposedData), &data); err != nil {
				logrus.Error("FAQ ApproveRequest unmarshal error:", err)
				return response.Error(http.StatusInternalServerError, "Failed to parse proposed data")
			}
		}

		faq := &gorm_model.FAQ{
			ID:       approval.RecordID,
			Question: data["question"],
			Answer:   data["answer"],
		}

		if err := u.gormDbRepo.CreateFAQ(ctx, faq); err != nil {
			logrus.Error("FAQ ApproveRequest create error:", err)
			return response.Error(http.StatusInternalServerError, "Failed to create FAQ")
		}

	case "UPDATE":
		existing, err := u.gormDbRepo.GetFAQByID(ctx, approval.RecordID)
		if err != nil {
			logrus.Error("FAQ ApproveRequest update fetch error:", err)
			return response.Error(http.StatusInternalServerError, "Failed to fetch FAQ for update")
		}

		var data map[string]string
		if approval.ProposedData != nil {
			if err := json.Unmarshal([]byte(*approval.ProposedData), &data); err != nil {
				logrus.Error("FAQ ApproveRequest unmarshal error:", err)
				return response.Error(http.StatusInternalServerError, "Failed to parse proposed data")
			}
		}

		existing.Question = data["question"]
		existing.Answer = data["answer"]

		if err := u.gormDbRepo.UpdateFAQ(ctx, existing); err != nil {
			logrus.Error("FAQ ApproveRequest update error:", err)
			return response.Error(http.StatusInternalServerError, "Failed to update FAQ")
		}

	default:
		return response.Error(http.StatusBadRequest, "Unknown action: "+approval.Action)
	}

	// Mark approval as APPROVED
	approval.Status = "APPROVED"
	approval.ReviewedBySuperadminID = &superadminID
	if err := u.gormDbRepo.UpdateApprovalRequest(ctx, approval); err != nil {
		logrus.Error("FAQ ApproveRequest status update error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to update approval status")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Approve", "FAQ", approval.RecordID, nil, true)
	return response.Success(approval.ToApprovalRequestResp())
}

// RejectRequest — Superadmin rejects a FAQ approval request.
func (u *appUsecase) RejectRequest(ctx context.Context, superadminID string, approvalID string, req request_model.ReviewApprovalRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	approval, err := u.gormDbRepo.GetApprovalRequestByID(ctx, approvalID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Approval request not found")
		}
		logrus.Error("FAQ RejectRequest fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch approval request")
	}

	if approval.TableName != "faqs" {
		return response.Error(http.StatusBadRequest, "This approval request is not for FAQs")
	}

	if approval.Status != "PENDING" {
		return response.Error(http.StatusBadRequest, "Approval request is not in PENDING status")
	}

	approval.Status = "REJECTED"
	approval.ReviewedBySuperadminID = &superadminID
	approval.RejectedReason = req.RejectedReason

	if err := u.gormDbRepo.UpdateApprovalRequest(ctx, approval); err != nil {
		logrus.Error("FAQ RejectRequest update error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to update approval status")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Reject", "FAQ", approval.RecordID, req, true)
	return response.Success(approval.ToApprovalRequestResp())
}

// ---------- Public — no auth ----------

// FetchPublic returns all FAQs from the DB (all are approved since they only enter after approval).
func (u *appUsecase) FetchPublic(ctx context.Context, page, limit int64, search *string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit
	filter := gorm_model.FAQFilter{
		Search: search,
	}
	filter.Limit = &limit
	filter.Offset = &offset
	if len(filter.Sorts) == 0 {
		filter.Sorts = []map[string]string{
			{"created_at": "DESC"},
		}
	}

	total, err := u.gormDbRepo.CountFAQ(ctx, gorm_model.FAQFilter{Search: search})
	if err != nil {
		logrus.Error("FAQ FetchPublic count error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to count FAQ records")
	}

	faqs, err := u.gormDbRepo.FetchFAQ(ctx, filter)
	if err != nil {
		logrus.Error("FAQ FetchPublic fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch FAQ data")
	}

	var results []interface{}
	for _, f := range faqs {
		results = append(results, f.ToFAQResp())
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

// FetchPublicByID returns a single FAQ by ID (public).
func (u *appUsecase) FetchPublicByID(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	faq, err := u.gormDbRepo.GetFAQByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "FAQ not found")
		}
		logrus.Error("FAQ FetchPublicByID error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch FAQ")
	}

	return response.Success(faq.ToFAQResp())
}
