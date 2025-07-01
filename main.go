package main

import (
	"log"
	"os"

	"github.com/Mahamudul-Dev/aisense_portal_snapshot/db"
	"github.com/Mahamudul-Dev/aisense_portal_snapshot/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("❌ Error loading .env file: %v", err)
	}
	db.Init()

	router := gin.Default()
	port := os.Getenv("PORT")

	api := router.Group("/api")
	{
		api.POST("/snapshots", handlers.CreateSnapshot)
		api.GET("/bucket/:name", handlers.RequestNewBucket)
		api.GET("/snapshots", handlers.AuthorizeSnapshot)
		api.POST("/snapshots/bulk", handlers.AuthorizeBulkSnapshots)
		api.DELETE("/bulk-objects", handlers.DeleteBulkObjects)
		api.DELETE("/objects", handlers.DeleteObject)

	}

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "App is running",
			"version":   "1.0.0",
			"author":    "Aisense USA",
			"developer": "Mahamudul Hasan",
			"port":      port,
			"application routes": gin.H{
				"create snapshot": gin.H{
					"method": "POST",
					"path":   "/api/snapshots",
					"body": gin.H{"image": "file", "device_id": "int", "captured_at": "time", "file_available": "bool", "device_name": "string",
						"detection": map[string]string{}},
					"type": "multipart/form-data",
				},
				"authorize snapshot": gin.H{
					"method": "GET",
					"path":   "/api/snapshots?uri=gs://bucket/object",
				},
				"authorize bulk snapshots": gin.H{
					"method": "POST",
					"path":   "/api/snapshots/bulk",
					"body":   gin.H{"uris": []string{}},
				},
				"delete object": gin.H{
					"method": "DELETE",
					"path":   "/api/snapshots",
					"body":   gin.H{"gcs_uri": "gs://bucket/object"},
				},
				"delete bulk objects": gin.H{
					"method": "DELETE",
					"path":   "/api/bulk-objects",
					"body":   gin.H{"gcs_uris": []string{}},
				},
				"create new bucket": gin.H{
					"method": "GET",
					"path":   "/api/bucket/:name",
					"note":   "Replace :name with user name",
				},
			},
		})
	})

	router.Run(":" + port)
}
