package gcs

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// CreateBucket creates a new GCS bucket in the given project and location.
func CreateBucket(projectID, bucketName, location string) error {
	ctx := context.Background()

	// Use your service account key JSON file
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(os.Getenv("SERVICE_ACCOUNT_JSON_FILE_PATH")))
	if err != nil {
		log.Panicln(err)
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	// Check if bucket already exists
	_, err = client.Bucket(bucketName).Attrs(ctx)
	if err == nil {
		return fmt.Errorf("bucket %s already exists", bucketName)
	}

	bucket := client.Bucket(bucketName)
	bucketAttrs := &storage.BucketAttrs{
		Location:                 location, // e.g. "US", "ASIA", "EU",
		UniformBucketLevelAccess: storage.UniformBucketLevelAccess{Enabled: true},
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if err := bucket.Create(ctx, projectID, bucketAttrs); err != nil {
		log.Panicln(err)
		return fmt.Errorf("failed to create bucket: %v", err)
	}

	fmt.Printf("✅ Bucket %s created successfully in %s\n", bucketName, location)
	return nil
}

// BucketExists checks if a bucket with the given name exists.
func BucketExists(bucketName string) (bool, error) {
	ctx := context.Background()

	println(bucketName)

	credentialPath := os.Getenv("SERVICE_ACCOUNT_JSON_FILE_PATH")

	println(credentialPath)

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialPath))
	if err != nil {
		return false, fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()

	_, err = client.Bucket(bucketName).Attrs(ctx)
	if err != nil {
		if err == storage.ErrBucketNotExist {
			return false, nil
		} else if e, ok := err.(*googleapi.Error); ok {
			if e.Code == 403 || e.Code == 404 {
				// Treat both 403 and 404 as "bucket doesn't exist" (to handle permission-limited environments)
				return false, nil
			}
		} else {
			return false, fmt.Errorf("error checking bucket: %v", err)
		}

	}

	return true, nil
}
