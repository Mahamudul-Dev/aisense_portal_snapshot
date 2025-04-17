package models

import (
	"time"
)

type Snapshot struct {
	ID               int       `gorm:"primaryKey;autoIncrement"`
	DeviceID         int       `json:"device_id"`
	RpiNo            string    `json:"rpi_no"`
	DistanceCM       float64   `json:"distance_cm"`
	ImagePath        string    `json:"image_path"`
	AuthenticatedURL string    `json:"authenticated_url"`
	CapturedAt       time.Time `gorm:"autoCreateTime"`
	PersonCount      int       `json:"person_count"`
	DeerCount        int       `json:"deer_count"`
	RaccoonCount     int       `json:"raccoon_count"`
	SquirrelCount    int       `json:"squirrel_count"`
	RabbitCount      int       `json:"rabbit_count"`
	DogCount         int       `json:"dog_count"`
	CatCount         int       `json:"cat_count"`
	CarCount         int       `json:"car_count"`
	FoxCount         int       `json:"fox_count"`
	CowCount         int       `json:"cow_count"`
	SheepCount       int       `json:"sheep_count"`
	HorseCount       int       `json:"horse_count"`
	FileAvailable    bool      `json:"file_available"`
}

type Device struct {
	ID           int       `gorm:"primaryKey;autoIncrement"`
	Name         string    `gorm:"not null"`
	Description  string    `gorm:"not null"`
	LastModified time.Time `gorm:"autoUpdateTime"`
	UserName     string    `gorm:"not null"`
	Password     string    `gorm:"not null"`
	DeviceUserID *int      `gorm:"column:device_user_id"`
	User         *User     `gorm:"foreignKey:DeviceUserID;references:ID"`
}

type User struct {
	ID           int    `gorm:"primaryKey;autoIncrement"`
	Name         string `gorm:"not null"`
	Email        string `gorm:"unique;not null"`
	Password     *string
	Role         string `gorm:"not null"`
	LastLogin    *time.Time
	LastLogout   *time.Time
	LastLoginIP  *string
	LastLogoutIP *string
	AuthToken    *string
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
