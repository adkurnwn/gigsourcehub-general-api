package usecase_cv

import (
	"context"
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
	uploadResp, err := u.storageRepo.UploadFilePublic(objectKey, file, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return response.Error(http.StatusInternalServerError, "upload failed: "+err.Error())
	}

	// 3. Save Metadata to DB
	cv := &gorm_model.CV{
		ID:     uuid.New().String(),
		UserID: userID,
		Name:   fileHeader.Filename,
		Path:   uploadResp.URL, // or objectKey
	}

	if err := u.gormRepo.CreateCV(ctx, cv); err != nil {
		return response.Error(http.StatusInternalServerError, "db save failed")
	}

	// 4. Publish Event
	go func() {
		// Use a detached context or background context for async publishing
		// to avoid cancellation if the request context is cancelled.
		// However, for simplicity here, we might just log validation errors.
		// A robust solution would use an outbox pattern or a separate worker.
		// Here we just fire and forget with a new context.
		bgCtx := context.Background()
		err := u.mqRepo.Publish(bgCtx, os.Getenv("RABBITMQ_QUEUE_CV_UPLOAD"), map[string]interface{}{
			"event":   "cv_uploaded",
			"user_id": userID,
			"cv_id":   cv.ID,
			"path":    cv.Path,
			"time":    time.Now(),
		})
		if err != nil {
			fmt.Printf("failed to publish message: %v\n", err)
		}
	}()

	return response.Success(cv.ToCVResp())
}
