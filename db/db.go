package db

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	db_name := os.Getenv("DB_NAME")
	portStr := os.Getenv("DB_PORT")
	port, er := strconv.Atoi(portStr)
	if er != nil {
		log.Fatalf("Invalid DB_PORT: %v", er)
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		host, user, password, db_name, port, "disable",
	)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		// NamingStrategy: schema.NamingStrategy{
		// 	SingularTable: true,
		// },
	})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	} else {
		log.Println("Connected to database")
	}
}
