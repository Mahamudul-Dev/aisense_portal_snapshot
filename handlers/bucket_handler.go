package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Mahamudul-Dev/aisense_portal_snapshot/gcs"
	"github.com/Mahamudul-Dev/aisense_portal_snapshot/utils"
	"github.com/gin-gonic/gin"
)

func RequestNewBucket(c *gin.Context) {
	userName, exist := c.Params.Get("name")

	if exist {

		bucketExist, bucketErr := gcs.BucketExists(userName)

		if bucketErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check bucket existence"})
			return
		}

		if bucketExist {
			c.JSON(http.StatusOK, gin.H{"message": "Bucket created successfully", "bucketName": userName})
			return
		}

		// Remove all spaces and convert to lowercase
		userName = strings.ToLower(strings.ReplaceAll(userName, " ", ""))
		fmt.Println(userName) // Output: jillianbright
	}

	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User name is required"})
		return
	}

	bucketName := userName + "_" + utils.GenerateUUID()

	if err := gcs.CreateBucket(os.Getenv("GCP_PROJECT_ID"), bucketName, "US"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bucket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bucket created successfully", "bucketName": bucketName})
}
