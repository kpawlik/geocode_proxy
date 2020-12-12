package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Authentication part for google service
type Authentication struct {
	APIKey       string `json:"apiKey"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Channel      string `json:"channel"`
}

// Config stores configuration
type Config struct {
	Authentication    Authentication `json:"authentication"`
	WorkersNumber     int            `json:"workersNumber"`
	RequestsPerSecond int            `json:"requestsPerSecond"`
	Port              int            `json:"port"`
	LogLevel          string         `json:"logLevel"`
}

// ReadConfig reads configuration from file
func ReadConfig(filepath string) (cfg Config, err error) {
	var (
		data []byte
		e    error
	)
	cfg = Config{LogLevel: "info", Port: 8888, WorkersNumber: 10}
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
