package utils

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Aws_access_key_id     string `json:"aws_access_key_id"`
	Aws_secret_access_key string `json:"aws_secret_access_key"`
	Region                string `json:"region"`
	SessionKey            string `json:"sessionKey"`
}

type Secret struct {
	Config `json:"Secrets"`
}

func LoadConfig() (Config, error) {
	config := Secret{}
	fd, err := os.Open("Client/config/config.json")
	if err != nil {
		return config.Config, err
	}
	defer fd.Close()
	data, err := io.ReadAll(fd)
	if err != nil {
		return config.Config, err
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config.Config, err
	}
	return config.Config, nil

}
