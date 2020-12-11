package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Config stores configuration
type Config struct {
	APIKey            string `json:"apiKey"`
	ClientID          string `json:"clientId"`
	ClientSecret      string `json:"clientSecret"`
	Channel           string `json:"channel"`
	WorkersNumber     int    `json:"workersNumber"`
	RequestsPerSecond int    `json:"requestsPerSecond"`
	Port              int    `json:"port"`
}

// ReadConfig reads configuration from file
func ReadConfig(filepath string) (cfg Config, err error) {
	var (
		data []byte
		e    error
	)
	if data, e = ioutil.ReadFile(filepath); err != nil {
		err = fmt.Errorf("Error reading file %s, %v", filepath, e)
		return
	}
	if e = json.Unmarshal(data, &cfg); e != nil {
		err = fmt.Errorf("Error unmarshal config data from %s, %v", filepath, e)
		return
	}
	return
}
