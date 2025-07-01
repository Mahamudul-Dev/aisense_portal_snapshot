package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
)

func GetPrivateKey(jsonPath string) []byte {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		panic("❌ Failed to read service account key file: " + err.Error())
	}
	var conf struct {
		PrivateKey string `json:"private_key"`
	}
	if err := json.Unmarshal(data, &conf); err != nil {
		panic("❌ Failed to parse service account key file: " + err.Error())
	}
	return []byte(conf.PrivateKey)
}

// GenerateRandomID returns a random 5-digit integer as string
func GenerateRandomID() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	id := r.Intn(90000) + 10000 // 10000 to 99999
	return fmt.Sprintf("%d", id)
}

func GenerateUUID() string {
	// Generate a new UUID (version 4)
	id := uuid.New()
	return id.String() // Return UUID as a string
}
