package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type Snapshot struct {
	ID               int             `gorm:"primaryKey;autoIncrement"`
	Name             string          `json:"name"`
	DeviceID         int             `json:"device_id"`
	RpiNo            string          `json:"rpi_no"`
	DistanceCM       decimal.Decimal `gorm:"column:distance_cm;type:numeric"`
	ImagePath        string          `json:"image_path"`
	AuthenticatedURL string          `json:"authenticated_url"`
	CapturedAt       time.Time       `gorm:"autoCreateTime"`
	Detection        datatypes.JSON  `gorm:"type:jsonb"`
	FileAvailable    bool            `json:"file_available"`
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
	Bucket       *string   `json:"bucket"`
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
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	Devices      []Device  `gorm:"foreignKey:DeviceUserID;references:ID"`
}

type Classes struct {
	ID          int    `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"unique;not null"`
	Description *string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

type ServiceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}
