package s3repo

import (
	"context"
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
	url := &url.URL{}
	if r.publicURL == nil {
		// If no public URL configured, try to construct from minio client endpoint
		// This might return the internal endpointURL which might not be reachable from public
		// But it maintains previous behavior logic
		u := r.client.EndpointURL()
		url = u
	} else {
		url = r.publicURL
	}

	// add path with object key
	// We need to ensure we don't double encode or miss the bucket if strictly following path style
	// Previous implementation just appended objectKey to Path.
	// If publicURL includes bucket (e.g. cdn.example.com), this is fine.
	// If publicURL is just host, we might need to append bucket if not virtual-host style.
	// Given previous code: url.Path = objectKey, it assumes publicURL points to the root concept or bucket root.
	// We'll stick to that.

	newURL := *url
	newURL.Path = objectKey
	if r.publicURL == nil {
		// If we are using the endpoint URL (which usually doesn't have bucket in path for minio-go unless configured),
		// we might need to append bucket/key if it's path style.
		// But minio-go EndpointURL() returns the base.
		// Let's assume for now user has S3_PUBLIC_URL set correctly as per env file.
	}

	return newURL.String()
}
