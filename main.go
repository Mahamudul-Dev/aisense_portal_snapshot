package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/Mahamudul-Dev/aisense_portal_snapshot/db"
	"github.com/Mahamudul-Dev/aisense_portal_snapshot/handlers"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("❌ Error loading .env file: %v", err)
	}
	db.Init()

	router := gin.Default()

	api := router.Group("/api")
	{
		api.POST("/snapshots", handlers.CreateSnapshot)
	}

	router.Run(":8080")
}
