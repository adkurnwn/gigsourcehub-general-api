package usecase_career_department

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)


// ---------- Admin & Superadmin — CMS ----------

// FetchAll returns all CareerDepartment records from the DB merged with pending/rejected approvals.
func (u *appUsecase) FetchAll(ctx context.Context, page, limit int64, search *string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1. Fetch approved career departments
	allApproved, err := u.gormDbRepo.FetchCareerDepartment(ctx, gorm_model.CareerDepartmentFilter{})
	if err != nil {
		logrus.Error("CareerDepartment FetchAll approved fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch career department data")
	}

	// 2. Fetch all approvals for career_departments
	tableName := "career_departments"
	approvals, err := u.gormDbRepo.FetchApprovalRequests(ctx, gorm_model.ApprovalRequestFilter{
		TableNameEq: &tableName,
	})
	if err != nil {
		logrus.Error("CareerDepartment FetchAll approvals fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch career department approvals")
	}

	// 3. Merge them
	type tempDept struct {
		ID          string
		Name        string
		Description string
		ImagePath   *string
		Author      string
		Status      string
		PublishedAt *time.Time
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}

	// Build map of approved departments by ID
	approvedMap := make(map[string]tempDept)
	for _, d := range allApproved {
		var pub *time.Time = &d.CreatedAt
		approvedMap[d.ID] = tempDept{
			ID:          d.ID,
			Name:        d.Name,
			Description: d.Description,
			ImagePath:   d.ImagePath,
			Status:      "PUBLISHED",
			PublishedAt: pub,
			CreatedAt:   d.CreatedAt,
			UpdatedAt:   d.UpdatedAt,
		}
	}

	// Build map of the latest approval request by RecordID
	latestApproval := make(map[string]gorm_model.ApprovalRequest)
	for _, app := range approvals {
		existing, exists := latestApproval[app.RecordID]
		if !exists || app.CreatedAt.After(existing.CreatedAt) {
			latestApproval[app.RecordID] = app
		}
	}

	// Create a set of all record IDs we need to process
	allRecordIDs := make(map[string]bool)
	for id := range approvedMap {
		allRecordIDs[id] = true
	}
	for id := range latestApproval {
		allRecordIDs[id] = true
	}

	finalMap := make(map[string]tempDept)

	for recordID := range allRecordIDs {
		dept, existsInDept := approvedMap[recordID]
		app, hasApproval := latestApproval[recordID]

		if existsInDept {
			// Department is active (published) in the database
			authorName := ""
			if hasApproval && app.RequestedByAdmin != nil {
				authorName = app.RequestedByAdmin.Name
			}

			if hasApproval && app.Status == "PENDING" {
				// There is a pending update
				var proposed map[string]string
				if app.ProposedData != nil {
					_ = json.Unmarshal([]byte(*app.ProposedData), &proposed)
				}

				finalMap[recordID] = tempDept{
					ID:          recordID,
					Name:        proposed["name"],
					Description: proposed["description"],
					ImagePath:   dept.ImagePath,
					Author:      authorName,
					Status:      "DRAFT",
					PublishedAt: dept.PublishedAt,
					CreatedAt:   dept.CreatedAt,
					UpdatedAt:   app.UpdatedAt,
				}
			} else {
				finalMap[recordID] = tempDept{
					ID:          recordID,
					Name:        dept.Name,
					Description: dept.Description,
					ImagePath:   dept.ImagePath,
					Author:      authorName,
					Status:      "PUBLISHED",
					PublishedAt: dept.PublishedAt,
					CreatedAt:   dept.CreatedAt,
					UpdatedAt:   dept.UpdatedAt,
				}
			}
		} else {
			// Department is NOT active in the database yet (pending create or rejected)
			if hasApproval && (app.Status == "PENDING" || app.Status == "REJECTED") && app.Action == "CREATE" {
				var proposed map[string]string
				if app.ProposedData != nil {
					_ = json.Unmarshal([]byte(*app.ProposedData), &proposed)
				}

				status := "DRAFT"
				if app.Status == "REJECTED" {
					status = "REJECTED"
				}

				authorName := ""
				if app.RequestedByAdmin != nil {
					authorName = app.RequestedByAdmin.Name
				}

				var imgPath *string
				if val, ok := proposed["image_path"]; ok && val != "" {
					imgPath = &val
				}

				finalMap[recordID] = tempDept{
					ID:          recordID,
					Name:        proposed["name"],
					Description: proposed["description"],
					ImagePath:   imgPath,
					Author:      authorName,
					Status:      status,
					PublishedAt: nil,
					CreatedAt:   app.CreatedAt,
					UpdatedAt:   app.UpdatedAt,
				}
			}
		}
	}

	// Convert map to slice and apply search filter
	var merged []tempDept
	for _, item := range finalMap {
		matchesSearch := true
		if search != nil && *search != "" {
			qLower := strings.ToLower(*search)
			matchesSearch = strings.Contains(strings.ToLower(item.Name), qLower) ||
				strings.Contains(strings.ToLower(item.Description), qLower)
		}

		if matchesSearch {
			merged = append(merged, item)
		}
	}

	// Sort merged slice by CreatedAt DESC
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].CreatedAt.After(merged[j].CreatedAt)
	})

	// Apply pagination (in-memory)
	total := int64(len(merged))
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	var paginated []tempDept
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		paginated = merged[offset:end]
	}

	// Convert to response shape
	var results []interface{}
	for _, item := range paginated {
		var authorPtr *string
		if item.Author != "" {
			authorPtr = &item.Author
		}
		statusVal := item.Status

		// Resolve image URL
		var imageURL *string
		if item.ImagePath != nil && *item.ImagePath != "" {
			ip := *item.ImagePath
			if len(ip) > 0 && ip[0] != 'h' {
				ip = fmt.Sprintf("%s/%s", getS3PublicURL(), ip)
			}
			imageURL = &ip
		}

		results = append(results, gorm_model.CareerDepartmentResp{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			ImageURL:    imageURL,
			Author:      authorPtr,
			Status:      &statusVal,
			PublishedAt: item.PublishedAt,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
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

// FetchData returns a single CareerDepartment by ID.
func (u *appUsecase) FetchData(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	dept, err := u.gormDbRepo.GetCareerDepartmentByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Check if there is a pending/rejected CREATE approval request for this record ID
			appReq, appErr := u.gormDbRepo.GetPendingApprovalByRecord(ctx, "career_departments", id)
			if appErr != nil {
				// Try by ApprovalRequest ID itself
				appReq2, appErr2 := u.gormDbRepo.GetApprovalRequestByID(ctx, id)
				if appErr2 == nil && appReq2 != nil && appReq2.TableName == "career_departments" && appReq2.Action == "CREATE" {
					appReq = appReq2
					id = appReq2.RecordID
				} else {
					return response.Error(http.StatusNotFound, "Career department not found")
				}
			}
			var data map[string]string
			if appReq.ProposedData != nil {
				_ = json.Unmarshal([]byte(*appReq.ProposedData), &data)
			}
			var imageURL *string
			if val, ok := data["image_path"]; ok && val != "" {
				ip := val
				if len(ip) > 0 && ip[0] != 'h' {
					ip = fmt.Sprintf("%s/%s", getS3PublicURL(), ip)
				}
				imageURL = &ip
			}
			statusVal := "DRAFT"
			if appReq.Status == "REJECTED" {
				statusVal = "REJECTED"
			}
			return response.Success(gorm_model.CareerDepartmentResp{
				ID:          id,
				Name:        data["name"],
				Description: data["description"],
				ImageURL:    imageURL,
				Status:      &statusVal,
				CreatedAt:   appReq.CreatedAt,
				UpdatedAt:   appReq.UpdatedAt,
			})
		}
		logrus.Error("CareerDepartment FetchData error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch career department")
	}

	status := "PUBLISHED"
	approvals, err := u.gormDbRepo.FetchApprovalRequests(ctx, gorm_model.ApprovalRequestFilter{
		RecordID: &id,
	})
	if err == nil && len(approvals) > 0 {
		var latest gorm_model.ApprovalRequest
		for _, app := range approvals {
			if latest.ID == "" || app.CreatedAt.After(latest.CreatedAt) {
				latest = app
			}
		}
		if latest.Status == "PENDING" {
			status = "DRAFT"
		} else if latest.Status == "REJECTED" {
			status = "REJECTED"
		}
	}

	return response.Success(dept.ToCareerDepartmentResp(status))
}

// ---------- Admin only ----------

// Create — Admin creates a CareerDepartment approval request (CREATE action).
// The actual record is NOT inserted into career_departments table yet.
func (u *appUsecase) Create(ctx context.Context, adminID string, req request_model.CreateCareerDepartmentRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Serialize proposed data
	proposedJSON, err := json.Marshal(map[string]string{
		"name":        req.Name,
		"description": req.Description,
	})
	if err != nil {
		logrus.Error("CareerDepartment Create marshal error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to process career department data")
	}
	proposedStr := string(proposedJSON)

	// Generate a future record ID for the career department
	recordID := uuid.New().String()

	approval := &gorm_model.ApprovalRequest{
		RequestedByAdminID: adminID,
		TableName:          "career_departments",
		RecordID:           recordID,
		Action:             "CREATE",
		ProposedData:       &proposedStr,
		Status:             "PENDING",
	}

	if err := u.gormDbRepo.CreateApprovalRequest(ctx, approval); err != nil {
		logrus.Error("CareerDepartment Create approval error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Create", "CareerDepartment", req.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create career department approval request")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Create", "CareerDepartment", req.Name, req, true)
	return response.Success(approval.ToApprovalRequestResp())
}

// Update — Admin creates a CareerDepartment approval request (UPDATE action).
func (u *appUsecase) Update(ctx context.Context, adminID string, id string, req request_model.UpdateCareerDepartmentRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Check if career department exists
	existing, err := u.gormDbRepo.GetCareerDepartmentByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Career department not found")
		}
		logrus.Error("CareerDepartment Update fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch career department")
	}

	// Serialize proposed data
	proposedJSON, err := json.Marshal(map[string]string{
		"name":        req.Name,
		"description": req.Description,
	})
	if err != nil {
		logrus.Error("CareerDepartment Update marshal error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to process career department data")
	}
	proposedStr := string(proposedJSON)

	approval := &gorm_model.ApprovalRequest{
		RequestedByAdminID: adminID,
		TableName:          "career_departments",
		RecordID:           existing.ID,
		Action:             "UPDATE",
		ProposedData:       &proposedStr,
		Status:             "PENDING",
	}

	if err := u.gormDbRepo.CreateApprovalRequest(ctx, approval); err != nil {
		logrus.Error("CareerDepartment Update approval error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Update", "CareerDepartment", existing.Name, req, false)
		return response.Error(http.StatusInternalServerError, "Failed to create career department update approval request")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Update", "CareerDepartment", existing.Name, req, true)
	return response.Success(approval.ToApprovalRequestResp())
}

// Delete — Admin directly soft-deletes a CareerDepartment (no approval needed).
func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	existing, err := u.gormDbRepo.GetCareerDepartmentByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Career department not found")
		}
		logrus.Error("CareerDepartment Delete fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch career department")
	}

	if err := u.gormDbRepo.DeleteCareerDepartment(ctx, id); err != nil {
		logrus.Error("CareerDepartment Delete error:", err)
		helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "CareerDepartment", existing.Name, nil, false)
		return response.Error(http.StatusInternalServerError, "Failed to delete career department")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Delete", "CareerDepartment", existing.Name, nil, true)
	return response.Success(nil)
}

// UploadImage — Admin uploads an image for a career department directly to S3.
// The image is stored under: career-department-images/{id}/filename.jpeg
// The objectKey (path) is saved to the career_departments record.
func (u *appUsecase) UploadImage(ctx context.Context, id string, fileHeader *multipart.FileHeader) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Check if career department exists
	dept, err := u.gormDbRepo.GetCareerDepartmentByID(ctx, id)
	var isPendingCreate bool
	var approval *gorm_model.ApprovalRequest
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			appReq, appErr := u.gormDbRepo.GetPendingApprovalByRecord(ctx, "career_departments", id)
			if appErr != nil {
				// Try by ApprovalRequest ID itself
				appReq2, appErr2 := u.gormDbRepo.GetApprovalRequestByID(ctx, id)
				if appErr2 == nil && appReq2 != nil && appReq2.TableName == "career_departments" && appReq2.Action == "CREATE" && appReq2.Status == "PENDING" {
					appReq = appReq2
					id = appReq2.RecordID
				} else {
					if appErr == gorm.ErrRecordNotFound {
						return response.Error(http.StatusNotFound, "Career department not found")
					}
					logrus.Error("CareerDepartment UploadImage approval fetch error:", appErr)
					return response.Error(http.StatusInternalServerError, "Failed to fetch career department status")
				}
			}
			isPendingCreate = true
			approval = appReq
		} else {
			logrus.Error("CareerDepartment UploadImage fetch error:", err)
			return response.Error(http.StatusInternalServerError, "Failed to fetch career department")
		}
	}

	// Validate content type
	contentType := fileHeader.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !allowedTypes[contentType] {
		return response.Error(http.StatusBadRequest, "Invalid file type. Only JPEG, PNG, and WebP are allowed")
	}

	// Read file
	file, err := fileHeader.Open()
	if err != nil {
		return response.Error(http.StatusBadRequest, "Failed to open file")
	}
	fileBytes, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		return response.Error(http.StatusBadRequest, "Failed to read file")
	}

	// Decode and re-encode as optimized JPEG
	originalImg, err := imaging.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return response.Error(http.StatusBadRequest, "Failed to decode image")
	}

	var originalBuf bytes.Buffer
	if err := imaging.Encode(&originalBuf, originalImg, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to optimize image")
	}

	// Upload to S3 under: career-department-images/{id}/{timestamp}.jpeg
	timestamp := time.Now().Unix()
	objectKey := fmt.Sprintf("career-department-images/%s/%d.jpeg", id, timestamp)

	var oldImagePath *string
	if isPendingCreate {
		var proposed map[string]string
		if approval.ProposedData != nil {
			_ = json.Unmarshal([]byte(*approval.ProposedData), &proposed)
			if path, ok := proposed["image_path"]; ok && path != "" {
				oldImagePath = &path
			}
		}
	} else {
		oldImagePath = dept.ImagePath
	}

	_, err = u.storageRepo.UploadFilePublic(objectKey, &originalBuf, "image/jpeg")
	if err != nil {
		logrus.Error("CareerDepartment UploadImage S3 error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to upload image")
	}

	// Update record or approval with new image path
	if isPendingCreate {
		var proposed map[string]string
		if approval.ProposedData != nil {
			_ = json.Unmarshal([]byte(*approval.ProposedData), &proposed)
		} else {
			proposed = make(map[string]string)
		}
		proposed["image_path"] = objectKey

		proposedBytes, marshalErr := json.Marshal(proposed)
		if marshalErr != nil {
			logrus.Error("CareerDepartment UploadImage marshal error:", marshalErr)
			return response.Error(http.StatusInternalServerError, "Failed to process image metadata")
		}
		proposedStr := string(proposedBytes)
		approval.ProposedData = &proposedStr

		if err := u.gormDbRepo.UpdateApprovalRequest(ctx, approval); err != nil {
			logrus.Error("CareerDepartment UploadImage approval update error:", err)
			return response.Error(http.StatusInternalServerError, "Failed to update career department image metadata")
		}
	} else {
		dept.ImagePath = &objectKey
		if err := u.gormDbRepo.UpdateCareerDepartment(ctx, dept); err != nil {
			logrus.Error("CareerDepartment UploadImage DB update error:", err)
			return response.Error(http.StatusInternalServerError, "Failed to update career department image")
		}
	}

	// Delete old image from S3 (best-effort)
	if oldImagePath != nil && *oldImagePath != "" && *oldImagePath != objectKey {
		if err := u.storageRepo.DeleteFile(*oldImagePath); err != nil {
			logrus.Warn("CareerDepartment UploadImage: failed to delete old image from S3:", err)
		}
	}

	imageURL := u.storageRepo.GetPublicLink(objectKey)
	deptName := ""
	if isPendingCreate {
		var proposed map[string]string
		if approval.ProposedData != nil {
			_ = json.Unmarshal([]byte(*approval.ProposedData), &proposed)
		}
		deptName = proposed["name"]
	} else {
		deptName = dept.Name
	}
	helpers.LogActivity(ctx, u.gormDbRepo, "UploadImage", "CareerDepartment", deptName, nil, true)
	return response.Success(map[string]string{
		"image_url": imageURL,
	})
}

// ---------- Superadmin — Approvals ----------

// FetchApprovals returns all approval requests for career departments.
func (u *appUsecase) FetchApprovals(ctx context.Context, page, limit int64) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit
	tableName := "career_departments"
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
		logrus.Error("CareerDepartment FetchApprovals count error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to count approval requests")
	}

	requests, err := u.gormDbRepo.FetchApprovalRequests(ctx, filter)
	if err != nil {
		logrus.Error("CareerDepartment FetchApprovals fetch error:", err)
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

// ApproveRequest — Superadmin approves a CareerDepartment approval request.
func (u *appUsecase) ApproveRequest(ctx context.Context, superadminID string, approvalID string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	approval, err := u.gormDbRepo.GetApprovalRequestByID(ctx, approvalID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Approval request not found")
		}
		logrus.Error("CareerDepartment ApproveRequest fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch approval request")
	}

	if approval.TableName != "career_departments" {
		return response.Error(http.StatusBadRequest, "This approval request is not for career departments")
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
				logrus.Error("CareerDepartment ApproveRequest unmarshal error:", err)
				return response.Error(http.StatusInternalServerError, "Failed to parse proposed data")
			}
		}

		var imagePath *string
		if val, ok := data["image_path"]; ok && val != "" {
			imagePath = &val
		}

		dept := &gorm_model.CareerDepartment{
			ID:          approval.RecordID,
			Name:        data["name"],
			Description: data["description"],
			ImagePath:   imagePath,
		}

		if err := u.gormDbRepo.CreateCareerDepartment(ctx, dept); err != nil {
			logrus.Error("CareerDepartment ApproveRequest create error:", err)
			return response.Error(http.StatusInternalServerError, "Failed to create career department")
		}

	case "UPDATE":
		existing, err := u.gormDbRepo.GetCareerDepartmentByID(ctx, approval.RecordID)
		if err != nil {
			logrus.Error("CareerDepartment ApproveRequest update fetch error:", err)
			return response.Error(http.StatusInternalServerError, "Failed to fetch career department for update")
		}

		var data map[string]string
		if approval.ProposedData != nil {
			if err := json.Unmarshal([]byte(*approval.ProposedData), &data); err != nil {
				logrus.Error("CareerDepartment ApproveRequest unmarshal error:", err)
				return response.Error(http.StatusInternalServerError, "Failed to parse proposed data")
			}
		}

		existing.Name = data["name"]
		existing.Description = data["description"]

		if err := u.gormDbRepo.UpdateCareerDepartment(ctx, existing); err != nil {
			logrus.Error("CareerDepartment ApproveRequest update error:", err)
			return response.Error(http.StatusInternalServerError, "Failed to update career department")
		}

	default:
		return response.Error(http.StatusBadRequest, "Unknown action: "+approval.Action)
	}

	// Mark approval as APPROVED
	approval.Status = "APPROVED"
	approval.ReviewedBySuperadminID = &superadminID
	if err := u.gormDbRepo.UpdateApprovalRequest(ctx, approval); err != nil {
		logrus.Error("CareerDepartment ApproveRequest status update error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to update approval status")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Approve", "CareerDepartment", approval.RecordID, nil, true)
	return response.Success(approval.ToApprovalRequestResp())
}

// RejectRequest — Superadmin rejects a CareerDepartment approval request.
func (u *appUsecase) RejectRequest(ctx context.Context, superadminID string, approvalID string, req request_model.ReviewApprovalRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	approval, err := u.gormDbRepo.GetApprovalRequestByID(ctx, approvalID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Approval request not found")
		}
		logrus.Error("CareerDepartment RejectRequest fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch approval request")
	}

	if approval.TableName != "career_departments" {
		return response.Error(http.StatusBadRequest, "This approval request is not for career departments")
	}

	if approval.Status != "PENDING" {
		return response.Error(http.StatusBadRequest, "Approval request is not in PENDING status")
	}

	approval.Status = "REJECTED"
	approval.ReviewedBySuperadminID = &superadminID
	approval.RejectedReason = req.RejectedReason

	if err := u.gormDbRepo.UpdateApprovalRequest(ctx, approval); err != nil {
		logrus.Error("CareerDepartment RejectRequest update error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to update approval status")
	}

	helpers.LogActivity(ctx, u.gormDbRepo, "Reject", "CareerDepartment", approval.RecordID, req, true)
	return response.Success(approval.ToApprovalRequestResp())
}

// ---------- Public — no auth ----------

// FetchPublic returns all approved career departments.
func (u *appUsecase) FetchPublic(ctx context.Context, page, limit int64, search *string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	offset := (page - 1) * limit
	filter := gorm_model.CareerDepartmentFilter{
		Search: search,
	}
	filter.Limit = &limit
	filter.Offset = &offset
	if len(filter.Sorts) == 0 {
		filter.Sorts = []map[string]string{
			{"created_at": "DESC"},
		}
	}

	total, err := u.gormDbRepo.CountCareerDepartment(ctx, gorm_model.CareerDepartmentFilter{Search: search})
	if err != nil {
		logrus.Error("CareerDepartment FetchPublic count error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to count career department records")
	}

	depts, err := u.gormDbRepo.FetchCareerDepartment(ctx, filter)
	if err != nil {
		logrus.Error("CareerDepartment FetchPublic fetch error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch career department data")
	}

	var results []interface{}
	for _, d := range depts {
		results = append(results, d.ToCareerDepartmentResp("PUBLISHED"))
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

// FetchPublicByID returns a single approved career department by ID.
func (u *appUsecase) FetchPublicByID(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	dept, err := u.gormDbRepo.GetCareerDepartmentByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.Error(http.StatusNotFound, "Career department not found")
		}
		logrus.Error("CareerDepartment FetchPublicByID error:", err)
		return response.Error(http.StatusInternalServerError, "Failed to fetch career department")
	}

	return response.Success(dept.ToCareerDepartmentResp("PUBLISHED"))
}

// ---------- helpers ----------

func getS3PublicURL() string {
	return os.Getenv("S3_PUBLIC_URL")
}
