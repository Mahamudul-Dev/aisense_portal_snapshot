package gcs

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"os"
)

func UploadFileAndGetGCSUri(bucketName, objectName, localFilePath string) (string, error) {
	ctx := context.Background()
	serviceAccountPath := os.Getenv("SERVICE_ACCOUNT_JSON_FILE_PATH")

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return "", fmt.Errorf("failed to create GCS client: %w", err)
	}
	defer client.Close()

	// Open local file
	f, err := os.Open(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open local file: %w", err)
	}
	defer f.Close()

	// Upload to GCS
	wc := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(wc, f); err != nil {
		return "", fmt.Errorf("failed to write to GCS: %w", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Return gsutil URI instead of signed URL
	gcsURI := fmt.Sprintf("gs://%s/%s", bucketName, objectName)
	return gcsURI, nil
}
