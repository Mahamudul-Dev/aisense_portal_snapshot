package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Mahamudul-Dev/aisense_portal_snapshot/db"
	"github.com/Mahamudul-Dev/aisense_portal_snapshot/gcs"
	"github.com/Mahamudul-Dev/aisense_portal_snapshot/models"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func CreateSnapshot(c *gin.Context) {
	// Parse form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	// Get uploaded file and its header
	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
		return
	}
	defer file.Close()

	// Parse device ID
	deviceID, err := strconv.Atoi(c.PostForm("device_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device_id"})
		return
	}

	// Parse detection JSON
	detectionRaw := c.PostForm("detection")
	var detectionData struct {
		Objects map[string]any `json:"objects"`
	}
	if err := json.Unmarshal([]byte(detectionRaw), &detectionData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid detection JSON"})
		return
	}

	// Load device with user
	var device models.Device
	if err := db.DB.Preload("User").First(&device, deviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	// Validate device name
	if c.PostForm("device_name") != device.Name {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device name does not match"})
		return
	}

	if device.Bucket == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device bucket is not set"})
		return
	}

	// Check bucket existence
	exists, err := gcs.BucketExists(*device.Bucket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check bucket existence"})
		return
	}
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bucket does not exist"})
		return
	}

	// Upload file directly from the multipart file stream
	imageURL, err := gcs.UploadFileAndGetGCSUriReader(*device.Bucket, fileHeader.Filename, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	// Start DB transaction
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		}
	}()

	// Validate detection objects
	objects := detectionData.Objects
	if objects == nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'objects' key in detection payload"})
		return
	}

	// Save detection classes
	for className := range objects {
		var existingClass models.Classes
		err := tx.Where("name = ?", className).First(&existingClass).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newClass := models.Classes{Name: className}
				if err := tx.Create(&newClass).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create class: " + err.Error()})
					return
				}
			} else {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
				return
			}
		}
	}

	// Marshal detection JSON to store in DB
	detectionJSON, err := json.Marshal(detectionData)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal detection"})
		return
	}

	// Create snapshot record
	snapshot := models.Snapshot{
		DeviceID:         deviceID,
		Name:             fileHeader.Filename,
		RpiNo:            device.Name,
		DistanceCM:       decimal.NewFromFloat(0.0),
		ImagePath:        "/" + *device.Bucket + "/" + fileHeader.Filename,
		AuthenticatedURL: imageURL,
		Detection:        detectionJSON,
		FileAvailable:    true,
	}

	if err := tx.Create(&snapshot).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create snapshot: " + err.Error()})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, snapshot)
}

// func CreateSnapshot(c *gin.Context) {
// 	// Parse form
// 	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
// 		return
// 	}

// 	// Get uploaded file
// 	file, fileHeader, err := c.Request.FormFile("image")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
// 		return
// 	}
// 	defer file.Close()

// 	// Parse form fields
// 	deviceID, err := strconv.Atoi(c.PostForm("device_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device_id"})
// 		return
// 	}
// 	detectionRaw := c.PostForm("detection") // string
// 	var detectionData struct {
// 		Objects map[string]any `json:"objects"`
// 	}

// 	if err := json.Unmarshal([]byte(detectionRaw), &detectionData); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid detection JSON"})
// 		return
// 	}

// 	var device models.Device

// 	// Load device with associated user
// 	if err := db.DB.Preload("User").First(&device, deviceID).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
// 		return
// 	}

// 	// Check device name
// 	if c.PostForm("device_name") != device.Name {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Device name does not match"})
// 		return
// 	}

// 	if device.Bucket == nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Device bucket is not set"})
// 		return
// 	}

// 	// Save file to temporary local file
// 	tempFileName := fmt.Sprintf("temp_%d_%s", time.Now().UnixNano(), fileHeader.Filename)
// 	localFilePath := filepath.Join(os.TempDir(), tempFileName)

// 	out, err := os.Create(localFilePath)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp file"})
// 		return
// 	}
// 	defer func() {
// 		out.Close()
// 		os.Remove(localFilePath) // Always clean up
// 	}()

// 	if _, err := io.Copy(out, file); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save temp image"})
// 		return
// 	}

// 	// Check bucket
// 	exist, err := gcs.BucketExists(*device.Bucket)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check bucket existence"})
// 		return
// 	}
// 	if !exist {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bucket does not exist"})
// 		return
// 	}

// 	// Upload file
// 	imageURL, err := gcs.UploadFileAndGetGCSUriReader(*device.Bucket, fileHeader.Filename, file)
// 	if err != nil {
// 		fmt.Println(err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
// 		return
// 	}

// 	// Start a DB transaction
// 	tx := db.DB.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 			fmt.Println(r)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
// 		}
// 	}()

// 	// Extract object classes (cars, people)
// 	objects := detectionData.Objects
// 	if objects == nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'objects' key in detection payload"})
// 		return
// 	}

// 	// Save detection classes
// 	for className := range objects {
// 		var existingClass models.Classes
// 		if err := tx.Where("name = ?", className).First(&existingClass).Error; err != nil {
// 			if errors.Is(err, gorm.ErrRecordNotFound) {
// 				newClass := models.Classes{Name: className}
// 				if err := tx.Create(&newClass).Error; err != nil {
// 					tx.Rollback()
// 					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create class: " + err.Error()})
// 					return
// 				}
// 			} else {
// 				tx.Rollback()
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
// 				return
// 			}
// 		}
// 	}

// 	// Marshal the full detection JSON again to store in DB
// 	detectionJSON, err := json.Marshal(detectionData)
// 	if err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal detection"})
// 		return
// 	}

// 	// Create snapshot
// 	snapshot := models.Snapshot{
// 		DeviceID:         deviceID,
// 		Name:             fileHeader.Filename,
// 		RpiNo:            device.Name,
// 		DistanceCM:       decimal.NewFromFloat(0.0),
// 		ImagePath:        "/" + *device.Bucket + "/" + fileHeader.Filename,
// 		AuthenticatedURL: imageURL,
// 		Detection:        detectionJSON,
// 		FileAvailable:    true,
// 	}

// 	if err := tx.Create(&snapshot).Error; err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create snapshot: " + err.Error()})
// 		return
// 	}

// 	// Commit transaction
// 	if err := tx.Commit().Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, snapshot)
// }

func AuthorizeSnapshot(c *gin.Context) {
	gsURI := c.Query("uri") // Expecting ?uri=gs://bucket/object

	if gsURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'uri' query parameter"})
		return
	}

	signedURL, err := gcs.GenerateSignedURL(gsURI, 30*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate signed URL", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"signed_url": signedURL,
	})
}

func AuthorizeBulkSnapshots(c *gin.Context) {
	var request struct {
		URIs []string `json:"uris"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	if len(request.URIs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No URIs provided"})
		return
	}

	signedURLs, err := gcs.GenerateBulkSignedURLs(request.URIs, 30*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate signed URLs", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"signed_urls": signedURLs,
	})
}

func DeleteObject(c *gin.Context) {
	var request struct {
		GCSUri string `json:"gcs_uri" binding:"required"`
	}

	// Parse JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Delete the object
	if err := gcs.DeleteObjectByURI(request.GCSUri); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete object: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Object deleted successfully"})
}

func DeleteBulkObjects(c *gin.Context) {
	var request struct {
		GCSUris []string `json:"gcs_uris" binding:"required"`
	}

	// Parse JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Delete the objects
	if err := gcs.DeleteBulkObjects(request.GCSUris); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete objects: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Objects deleted successfully"})
}
