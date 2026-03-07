package s3repo

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

func (r *s3Repo) GetPresignedLink(objectKey string, expires *time.Duration) string {
	expiry := time.Hour * 24 // default
	if expires != nil {
		expiry = *expires
	}

	// reqParams can be nil
	url, err := r.client.PresignedGetObject(context.TODO(), r.bucketName, objectKey, expiry, nil)
	if err != nil {
		logrus.Error("GetPresignedLink error: ", err)
		return ""
	}

	return url.String()
}

func (r *s3Repo) GetPublicLink(objectKey string) string {
	baseURL := &url.URL{}
	if r.publicURL == nil {
		u := r.client.EndpointURL()
		baseURL = u
	} else {
		baseURL = r.publicURL
	}

	newURL := *baseURL
	newURL.Path = fmt.Sprintf("/%s/%s", r.bucketName, objectKey)

	return newURL.String()
}
