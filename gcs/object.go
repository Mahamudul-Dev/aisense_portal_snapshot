package gcs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Mahamudul-Dev/aisense_portal_snapshot/models"
	"google.golang.org/api/option"
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

func UploadFileAndGetGCSUriReader(bucketName, objectName string, r io.Reader) (string, error) {
	ctx := context.Background()
	serviceAccountPath := os.Getenv("SERVICE_ACCOUNT_JSON_FILE_PATH")

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return "", fmt.Errorf("failed to create GCS client: %w", err)
	}
	defer client.Close()

	wc := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(wc, r); err != nil {
		return "", fmt.Errorf("failed to write to GCS: %w", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	gcsURI := fmt.Sprintf("gs://%s/%s", bucketName, objectName)
	return gcsURI, nil
}

// GenerateSignedURL generates a signed URL from a gs://bucket/object URI
func GenerateSignedURL(gsURI string, expiryDuration time.Duration) (string, error) {
	// Parse the gsutil URI
	if !strings.HasPrefix(gsURI, "gs://") {
		return "", fmt.Errorf("invalid gsutil URI: %s", gsURI)
	}

	uriParts := strings.SplitN(gsURI[len("gs://"):], "/", 2)
	if len(uriParts) != 2 {
		return "", fmt.Errorf("invalid gsutil URI format: %s", gsURI)
	}
	bucketName := uriParts[0]
	objectName, err := url.PathUnescape(uriParts[1])
	if err != nil {
		return "", fmt.Errorf("failed to unescape object name: %w", err)
	}

	// Load service account credentials
	serviceAccountPath := os.Getenv("SERVICE_ACCOUNT_JSON_FILE_PATH")

	saBytes, err := os.ReadFile(serviceAccountPath)
	if err != nil {
		return "", fmt.Errorf("failed to read service account file: %w", err)
	}

	var sa models.ServiceAccount
	if err := json.Unmarshal(saBytes, &sa); err != nil {
		return "", fmt.Errorf("failed to parse service account JSON: %w", err)
	}

	// Initialize storage client
	client, err := storage.NewClient(context.Background(), option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	// Generate signed URL
	opts := &storage.SignedURLOptions{
		GoogleAccessID: sa.ClientEmail,                                        // Make sure you set this env var
		PrivateKey:     []byte(strings.ReplaceAll(sa.PrivateKey, `\n`, "\n")), // Or load key file content properly
		Method:         "GET",
		Expires:        time.Now().Add(expiryDuration),
		Scheme:         storage.SigningSchemeV4,
	}

	signedURL, err := storage.SignedURL(bucketName, objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return signedURL, nil
}

// GenerateBulkSignedURLs generates signed URLs for a list of gs:// URIs
func GenerateBulkSignedURLs(gsURIs []string, expiryDuration time.Duration) (map[string]string, error) {
	results := make(map[string]string)

	for _, gsURI := range gsURIs {
		signedURL, err := GenerateSignedURL(gsURI, expiryDuration)
		if err != nil {
			return nil, fmt.Errorf("failed to generate signed URL for %s: %w", gsURI, err)
		}
		results[gsURI] = signedURL
	}

	return results, nil
}

func DeleteObjectByURI(gsURI string) error {
	// Parse gs://bucket-name/object-name
	if !strings.HasPrefix(gsURI, "gs://") {
		return fmt.Errorf("invalid GCS URI: %s", gsURI)
	}

	trimmed := strings.TrimPrefix(gsURI, "gs://")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid GCS URI format: %s", gsURI)
	}

	bucketName := parts[0]
	objectName := parts[1]

	ctx := context.Background()
	serviceAccountPath := os.Getenv("SERVICE_ACCOUNT_JSON_FILE_PATH")

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return fmt.Errorf("failed to create GCS client: %w", err)
	}
	defer client.Close()

	objectHandle := client.Bucket(bucketName).Object(objectName)

	if err := objectHandle.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

func parseGCSURI(gsURI string) (bucket, object string, err error) {
	if !strings.HasPrefix(gsURI, "gs://") {
		return "", "", fmt.Errorf("invalid GCS URI: %s", gsURI)
	}
	trimmed := strings.TrimPrefix(gsURI, "gs://")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GCS URI format: %s", gsURI)
	}
	return parts[0], parts[1], nil
}

func DeleteBulkObjects(gsURIs []string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(os.Getenv("SERVICE_ACCOUNT_JSON_FILE_PATH")))
	if err != nil {
		return fmt.Errorf("failed to create GCS client: %w", err)
	}
	defer client.Close()

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrency to 10
	var mu sync.Mutex
	var errs []error

	for _, uri := range gsURIs {
		bucket, object, err := parseGCSURI(uri)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore
		go func(bucket, object, uri string) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			if err := client.Bucket(bucket).Object(object).Delete(ctx); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to delete %s: %w", uri, err))
				mu.Unlock()
			}
		}(bucket, object, uri)
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("some deletions failed: %v", errs)
	}

	return nil
}
