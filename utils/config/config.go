package config

import (
	logging "botDetection/utils/logger"
	"encoding/json"
	"io/ioutil"
	"os"
)

type Credentials struct {
	Private `json:"Credentials"`
}

type Private struct {
	ACCESS_KEY string `json:"ACCESS_KEY"`
	SECRET_KEY string `json:"SECRET_KEY"`
}

func GetCredentials(filename string) map[string]interface{} {
	fd, err := os.Open(filename)
	if err != nil {
		logging.Fatal("ERROR opening configuration file", filename, "-", err)
	}
	defer fd.Close()
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		logging.Fatal("ERROR reading file: ", filename, "-", err)
	}

	var json_data map[string]interface{}

	err = json.Unmarshal([]byte(data), &json_data)
	if err != nil {
		logging.Fatal("Could not parse json configuration file: ", err)
	}

	return json_data

}
