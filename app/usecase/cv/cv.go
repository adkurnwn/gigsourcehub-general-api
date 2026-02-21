package usecase_cv

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/google/uuid"
)

type CVUsecase interface {
	UploadCV(ctx context.Context, userID string, fileHeader *multipart.FileHeader) response.Base
	GetParsedCV(ctx context.Context, cvID string) response.Base
	ConfirmCV(ctx context.Context, userID, cvID string, editedData map[string]interface{}) response.Base
}

type cvUsecase struct {
	gormRepo    domain.GormRepo
	storageRepo domain.StorageRepo
	mqRepo      domain.MessageBroker // <--- Inject MessageBroker
	timeout     time.Duration
}

func NewCVUsecase(gormRepo domain.GormRepo, storageRepo domain.StorageRepo, mqRepo domain.MessageBroker, timeout time.Duration) CVUsecase {
	return &cvUsecase{
		gormRepo:    gormRepo,
		storageRepo: storageRepo,
		mqRepo:      mqRepo,
		timeout:     timeout,
	}
}

func (u *cvUsecase) UploadCV(ctx context.Context, userID string, fileHeader *multipart.FileHeader) response.Base {
	// 1. Open File
	file, err := fileHeader.Open()
	if err != nil {
		return response.Error(http.StatusBadRequest, "failed to open file")
	}
	defer file.Close()

	// 2. Upload to Storage (S3/Minio)
	// Generate a unique key, e.g., "cvs/userID/filename.pdf"
	objectKey := fmt.Sprintf("cvs/%s/%s", userID, fileHeader.Filename)
	_, err = u.storageRepo.UploadFilePublic(objectKey, file, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return response.Error(http.StatusInternalServerError, "upload failed: "+err.Error())
	}

	// 3. Save Metadata to DB
	cv := &gorm_model.CV{
		ID:     uuid.New().String(),
		UserID: userID,
		Name:   fileHeader.Filename,
		Path:   objectKey, // Store relative path (e.g., cvs/uuid/file.pdf)
		// ParsedData is omitted so it correctly inserts NULL
	}

	if err := u.gormRepo.CreateCV(ctx, cv); err != nil {
		return response.Error(http.StatusInternalServerError, "db save failed")
	}

	// 4. Publish Event
	go func() {
		// Use a detached context or background context for async publishing
		// to avoid cancellation if the request context is cancelled.

		if u.mqRepo == nil {
			fmt.Println("mqRepo is nil, skipping event publishing")
			return
		}

		bgCtx := context.Background()
		// Here we just fire and forget with a new context.
		err := u.mqRepo.Publish(bgCtx, os.Getenv("RABBITMQ_QUEUE_CV_UPLOAD"), map[string]interface{}{
			"event":       "cv_uploaded",
			"user_id":     userID,
			"cv_id":       cv.ID,
			"path":        cv.Path,
			"uploaded_at": time.Now(),
		})
		if err != nil {
			fmt.Printf("failed to publish message: %v\n", err)
		}
	}()

	return response.Success(cv.ToCVResp())
}

func (u *cvUsecase) GetParsedCV(ctx context.Context, cvID string) response.Base {
	cv, err := u.gormRepo.GetCVByID(ctx, cvID)
	if err != nil {
		return response.Error(http.StatusNotFound, "cv not found")
	}

	return response.Success(cv.ToCVResp())
}

func (u *cvUsecase) ConfirmCV(ctx context.Context, userID, cvID string, editedData map[string]interface{}) response.Base {
	cv, err := u.gormRepo.GetCVByID(ctx, cvID)
	if err != nil {
		return response.Error(http.StatusNotFound, "cv not found")
	}

	if cv.UserID != userID {
		return response.Error(http.StatusForbidden, "not authorized to confirm this cv")
	}

	if cv.Status != "PARSED" {
		return response.Error(http.StatusBadRequest, "cv is not in PARSED status")
	}

	editedJSON, err := json.Marshal(editedData)
	if err != nil {
		return response.Error(http.StatusBadRequest, "invalid edited data")
	}

	jsonStr := string(editedJSON)
	cv.ParsedData = &jsonStr
	cv.Status = "CONFIRMED"

	if err := u.gormRepo.UpdateCV(ctx, cv); err != nil {
		return response.Error(http.StatusInternalServerError, "failed to update cv")
	}

	if u.mqRepo != nil {
		go func() {
			bgCtx := context.Background()
			_ = u.mqRepo.Publish(bgCtx, "final_cv", cv) // Use final_cv queue
		}()
	}

	return response.Success(cv.ToCVResp())
}
