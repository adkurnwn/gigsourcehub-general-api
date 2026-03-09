package s3repo

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

func (r *s3Repo) DeleteFile(objectKey string) error {
	err := r.client.RemoveObject(context.TODO(), r.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		logrus.Errorf("DeleteFile: Couldn't delete %v:%v. Here's why: %v\n", r.bucketName, objectKey, err)
		return err
	}
	return nil
}
