package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

type Config struct {
	Aws_access_key_id     string `json:"aws_access_key_id"`
	Aws_secret_access_key string `json:"aws_secret_access_key"`
	Region                string `json:"region"`
	Token                 string `json:"token"`
	SecretKey             string `json:"secretKey"`
}

var DataConfig = Config{}

func LoadEnv() error {
	configMap := map[string]string{}
	values := reflect.ValueOf(DataConfig)
	types := values.Type()
	for i := 0; i < values.NumField(); i++ {
		name := types.Field(i).Name
		secret, ok := os.LookupEnv(name)
		if !ok {
			return fmt.Errorf("could not resolve key: %s", name)
		}
		configMap[name] = secret
	}
	data, err := json.Marshal(configMap)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &DataConfig)
	if err != nil {
		return err
	}
	return nil

}

func (c *Config) GetSecret() string {
	return c.SecretKey
}

func (c *Config) GetRegion() string {
	return c.Region
}

func (c *Config) GetAwsAccess() string {
	return c.Aws_access_key_id
}

func (c *Config) GetAwsSecret() string {
	return c.Aws_secret_access_key
}
