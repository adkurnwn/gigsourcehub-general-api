package usecase_member

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
)

func (u *appUsecase) UploadProfilePicture(ctx context.Context, userID string, fileHeader *multipart.FileHeader) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

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

	// Fetch user first to get old profile picture key
	user, err := u.gormDbRepo.FetchOneUser(ctx, gorm_model.UserFilter{
		DefaultFilter: gorm_model.DefaultFilter{ID: userID},
	})
	if err != nil || user == nil {
		return response.Error(http.StatusNotFound, "User not found")
	}

	oldProfilePicture := user.ProfilePicture

	// Read file into buffer (needed for both S3 upload and async thumbnail)
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

	// Upload original (as optimized JPEG) to S3
	safeName := strings.ReplaceAll(user.Name, " ", "_")
	timestamp := time.Now().Unix()
	objectKey := fmt.Sprintf("profile-pictures/%s/%s_%d.jpeg", userID, safeName, timestamp)
	_, err = u.storageRepo.UploadFilePublic(objectKey, &originalBuf, "image/jpeg")
	if err != nil {
		logrus.Error("UploadProfilePicture S3 error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to upload profile picture")
	}

	// Update user record
	user.ProfilePicture = &objectKey
	if err := u.gormDbRepo.UpdateUser(ctx, user); err != nil {
		logrus.Error("UploadProfilePicture DB error: ", err)
		return response.Error(http.StatusInternalServerError, "Failed to update profile picture")
	}

	// Delete old profile picture + thumbnail from S3
	if oldProfilePicture != nil && *oldProfilePicture != "" && *oldProfilePicture != objectKey {
		if err := u.storageRepo.DeleteFile(*oldProfilePicture); err != nil {
			logrus.Warn("Failed to delete old profile picture from S3: ", err)
		}
		oldThumbKey := strings.TrimSuffix(*oldProfilePicture, filepath.Ext(*oldProfilePicture)) + "_thumb.jpeg"
		if err := u.storageRepo.DeleteFile(oldThumbKey); err != nil {
			logrus.Warn("Failed to delete old thumbnail from S3: ", err)
		}
	}

	// Async: generate JPEG thumbnail and upload
	thumbKey := fmt.Sprintf("profile-pictures/%s/%s_%d_thumb.jpeg", userID, safeName, timestamp)
	go func(imgBytes []byte, thumbObjectKey string) {
		img, err := imaging.Decode(bytes.NewReader(imgBytes))
		if err != nil {
			logrus.Warn("Thumbnail decode error: ", err)
			return
		}

		// Resize to 200px width, preserve aspect ratio
		thumb := imaging.Resize(img, 200, 0, imaging.Lanczos)

		// Encode as JPEG
		var buf bytes.Buffer
		if err := imaging.Encode(&buf, thumb, imaging.JPEG, imaging.JPEGQuality(80)); err != nil {
			logrus.Warn("Thumbnail encode error: ", err)
			return
		}

		// Upload thumbnail to S3
		if _, err := u.storageRepo.UploadFilePublic(thumbObjectKey, &buf, "image/jpeg"); err != nil {
			logrus.Warn("Thumbnail upload error: ", err)
		}
	}(fileBytes, thumbKey)

	return response.Success(map[string]string{
		"profile_picture": u.storageRepo.GetPublicLink(objectKey),
	})
}
