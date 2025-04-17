package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Mahamudul-Dev/aisense_portal_snapshot/db"
	"github.com/Mahamudul-Dev/aisense_portal_snapshot/gcs"
	"github.com/Mahamudul-Dev/aisense_portal_snapshot/models"
	"github.com/gin-gonic/gin"
)

func CreateSnapshot(c *gin.Context) {
	// Parse form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	// Get uploaded file
	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
		return
	}
	defer file.Close()

	exists, err := gcs.BucketExists(c.PostForm("rpi_no"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bucket"})
		return
	}

	if !exists {
		err := gcs.CreateBucket(os.Getenv("GCP_PROJECT_ID"), c.PostForm("rpi_no"), "us")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new bucket"})
			return
		}
	}

	// Save file to a temporary local file
	tempFileName := fmt.Sprintf("temp_%d_%s", time.Now().UnixNano(), fileHeader.Filename)
	localFilePath := filepath.Join(os.TempDir(), tempFileName)

	out, err := os.Create(localFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp file"})
		return
	}
	defer os.Remove(localFilePath) // Clean up temp file after upload
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save temp image"})
		return
	}

	// Parse form fields

	deviceID, _ := strconv.Atoi(c.PostForm("device_id"))
	distanceCM, _ := strconv.ParseFloat(c.PostForm("distance_cm"), 64)

	var device models.Device

	// Load device with associated user
	if err := db.DB.Preload("User").First(&device, deviceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	rpiNo := c.PostForm("rpi_no")
	deviceName := device.Name // assuming you've already fetched the device from DB

	// Get file extension
	fileExt := filepath.Ext(fileHeader.Filename) // e.g., ".jpg"

	// Format current time
	timestamp := time.Now().Format("20060102_150405") // yyyyMMdd_HHmmss

	// Create object name
	objectName := fmt.Sprintf("%s_%s_%s%s", rpiNo, deviceName, timestamp, fileExt)

	// Upload to GCS and get gs:// URI
	imageURL, err := gcs.UploadFileAndGetGCSUri(device.User.Name, objectName, localFilePath)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	snapshot := models.Snapshot{
		DeviceID:         deviceID,
		RpiNo:            rpiNo,
		DistanceCM:       distanceCM,
		ImagePath:        imageURL,
		AuthenticatedURL: imageURL,
		PersonCount:      atoi(c.PostForm("person_count")),
		DeerCount:        atoi(c.PostForm("deer_count")),
		RaccoonCount:     atoi(c.PostForm("raccoon_count")),
		SquirrelCount:    atoi(c.PostForm("squirrel_count")),
		RabbitCount:      atoi(c.PostForm("rabbit_count")),
		DogCount:         atoi(c.PostForm("dog_count")),
		CatCount:         atoi(c.PostForm("cat_count")),
		CarCount:         atoi(c.PostForm("car_count")),
		FoxCount:         atoi(c.PostForm("fox_count")),
		CowCount:         atoi(c.PostForm("cow_count")),
		SheepCount:       atoi(c.PostForm("sheep_count")),
		HorseCount:       atoi(c.PostForm("horse_count")),
		FileAvailable:    true,
	}

	if err := db.DB.Create(&snapshot).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save snapshot"})
		return
	}

	c.JSON(http.StatusCreated, snapshot)
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
