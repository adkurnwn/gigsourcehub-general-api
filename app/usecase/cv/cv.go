package usecase_cv

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
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
	storageRepo domain.StorageRepo // <--- Inject Storage Repo
	timeout     time.Duration
}

func NewCVUsecase(gormRepo domain.GormRepo, storageRepo domain.StorageRepo, timeout time.Duration) CVUsecase {
	return &cvUsecase{
		gormRepo:    gormRepo,
		storageRepo: storageRepo,
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

	return response.Success(cv.ToCVResp())
}
