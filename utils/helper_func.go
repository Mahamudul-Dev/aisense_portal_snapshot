package utils

import (
	"encoding/json"
	"os"
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
