package s3repo

import (
	"context"
	"io"
	"time"

	storage_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/storage"

	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

func (r *s3Repo) UploadFilePublic(objectKey string, body io.Reader, contentType string) (uploadData *storage_model.UploadResponse, err error) {
	// objectSize is -1 for unknown size, minio-go handles it (but limited to 5TB single part or requires ReadSeeke for multipart if size unknown? actually PutObject works withReader but performs better if size is known.
	// For simple migration we use -1. If issues arise we can read into buffer or require size.
	// However, minio-go PutObject with stream (io.Reader) and size -1 uploads effectively.

	_, err = r.client.PutObject(context.TODO(), r.bucketName, objectKey, body, -1, minio.PutObjectOptions{
		ContentType: contentType,
		// We can't strictly set ACL public-read here across all S3 implementations reliably via SDK options
		// without potential issues on some limited S3 providers or MinIO configurations.
		// Usually public access is managed by bucket policy.
		// If needed, we could try adding UserMetadata: map[string]string{"x-amz-acl": "public-read"}
	})
	if err != nil {
		logrus.Errorf(
			"UploadFilePublic: Couldn't upload file to %v:%v. Here's why: %v\n",
			r.bucketName, objectKey, err,
		)
		return
	}

	url := r.GetPublicLink(objectKey)
	if url == "" {
		logrus.Error("UploadFilePublic: GetPublicLink error")
		return
	}

	uploadData = &storage_model.UploadResponse{
		Key:         objectKey,
		ContentType: contentType,
		URL:         url,
	}

	return
}

func (r *s3Repo) UploadFilePrivate(objectKey string, body io.Reader, contentType string, expires *time.Duration) (uploadData *storage_model.UploadResponse, err error) {
	_, err = r.client.PutObject(context.TODO(), r.bucketName, objectKey, body, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		logrus.Errorf(
			"UploadFilePrivate: Couldn't upload file to %v:%v. Here's why: %v\n",
			r.bucketName, objectKey, err,
		)
		return
	}

	url := r.GetPresignedLink(objectKey, expires)
	if url == "" {
		logrus.Error("UploadFilePrivate: GetPresignedLink error")
		return
	}

	uploadData = &storage_model.UploadResponse{
		Key:         objectKey,
		ContentType: contentType,
		URL:         url,
	}

	return
}
