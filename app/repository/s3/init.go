package s3repo

import (
	"context"
	"net/url"
	"os"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type s3Repo struct {
	bucketName string
	publicURL  *url.URL
	client     *minio.Client
}

func NewS3Repo() domain.StorageRepo {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKeyID := os.Getenv("S3_ACCESS_KEY")
	secretAccessKey := os.Getenv("S3_SECRET_KEY")
	bucketName := os.Getenv("S3_BUCKET_NAME")
	useSSL := true

	// Parse endpoint to check for scheme and strip it if necessary for minio.New
	u, err := url.Parse(endpoint)
	if err == nil {
		if u.Scheme == "http" {
			useSSL = false
		}
		// minio.New expects endpoint without scheme
		endpoint = u.Host
	}

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		logrus.Fatalln(err)
	}

	// Check to see if we already own this bucket
	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		logrus.Warnf("s3: check bucket exists error: %v", err)
	} else if !exists {
		logrus.Infof("s3: bucket %s not found, creating...", bucketName)
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			logrus.Errorf("s3: failed to create bucket %s: %v", bucketName, err)
		} else {
			logrus.Infof("s3: bucket %s created", bucketName)
		}
	}

	publicURL, err := url.Parse(os.Getenv("S3_PUBLIC_URL"))
	if err != nil {
		logrus.Info("s3: without public url", err)
	}

	return &s3Repo{
		bucketName: bucketName,
		publicURL:  publicURL,
		client:     minioClient,
	}
}
