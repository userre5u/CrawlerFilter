package utils

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Secret   `json:"Secrets"`
	Database DB `json:"DB"`
}

type Secret struct {
	Aws_access_key_id     string `json:"aws_access_key_id"`
	Aws_secret_access_key string `json:"aws_secret_access_key"`
	Region                string `json:"region"`
	SessionKey            string `json:"sessionKey"`
}

type DB struct {
	User     string `json:"User"`
	Password string `json:"Password"`
	Port     string `json:"Port"`
}

func LoadConfig() (Config, error) {
	config := Config{}
	fd, err := os.Open("Client/config/config.json")
	if err != nil {
		return config, err
	}
	defer fd.Close()
	data, err := io.ReadAll(fd)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	return config, nil

}
