package system_setting

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/sirupsen/logrus"
)

func (u *appUsecase) Fetch(ctx context.Context) (*gorm_model.SystemSetting, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	data, err := u.repo.GetSystemSetting(ctx)
	if err != nil {
		return nil, err
	}

	if data.CVTemplatePath != nil && *data.CVTemplatePath != "" {
		data.CVTemplateURL = u.storageRepo.GetPublicLink(*data.CVTemplatePath)
	}

	return data, nil
}

func (u *appUsecase) Update(ctx context.Context, req request_model.UpdateSystemSettingRequest) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	model, err := u.repo.GetSystemSetting(ctx)
	if err != nil {
		return err
	}

	if req.IsAIModeEnabled != nil {
		model.IsAIModeEnabled = *req.IsAIModeEnabled
	}

	return u.repo.UpdateSystemSetting(ctx, model)
}

func (u *appUsecase) UploadCVTemplate(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1. Validate file extension (only .docx allowed)
	if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".docx") {
		return "", fmt.Errorf("only .docx files are allowed")
	}

	// 2. Open and read file
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file")
	}
	defer file.Close()

	// 3. Upload to S3 public folder under object key: public/cv-templates/CV_Template_[timestamp].docx
	timestamp := time.Now().Unix()
	objectKey := fmt.Sprintf("public/cv-templates/CV_Template_%d.docx", timestamp)

	// Fetch existing system settings to delete old template if it exists
	setting, err := u.repo.GetSystemSetting(ctx)
	if err != nil {
		return "", err
	}
	oldPath := setting.CVTemplatePath

	// Upload file (GHO uses application/vnd.openxmlformats-officedocument.wordprocessingml.document for docx files)
	_, err = u.storageRepo.UploadFilePublic(objectKey, file, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	if err != nil {
		return "", fmt.Errorf("failed to upload CV template to S3: %w", err)
	}

	// 4. Update system setting record
	setting.CVTemplatePath = &objectKey
	err = u.repo.UpdateSystemSetting(ctx, setting)
	if err != nil {
		// Best-effort: try to clean up newly uploaded file if DB update fails
		_ = u.storageRepo.DeleteFile(objectKey)
		return "", fmt.Errorf("failed to update system settings in DB: %w", err)
	}

	// 5. Delete old template from S3 (best-effort)
	if oldPath != nil && *oldPath != "" && *oldPath != objectKey {
		if err := u.storageRepo.DeleteFile(*oldPath); err != nil {
			logrus.Warn("UploadCVTemplate: failed to delete old CV template from S3:", err)
		}
	}

	return u.storageRepo.GetPublicLink(objectKey), nil
}
