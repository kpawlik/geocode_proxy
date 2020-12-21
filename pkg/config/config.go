package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

const (
	defaultPort          = 8888
	defaultLogLevel      = "info"
	defaultWorkersNumber = 10
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
	Quota             int            `json:"quota"`
	RequestPerMinute  int            `json:"requestPerMinute"`
	usedQuota         int
	mux               sync.RWMutex
}

func newConfig() *Config {
	cfg := &Config{LogLevel: defaultLogLevel,
		Port:          defaultPort,
		WorkersNumber: defaultWorkersNumber,
		mux:           sync.RWMutex{},
	}
	return cfg
}

// SetUsedQuota set value of the used  quota
func (c *Config) SetUsedQuota(q int) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.usedQuota = q
}

// IncQuota increments the value of used quota
func (c *Config) IncQuota() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.usedQuota++
}

// ResetUsedQuota set quota to default value
func (c *Config) ResetUsedQuota() {
	c.SetUsedQuota(c.Quota)
}

// CheckQuotaLimit checks if used quota exceeded
func (c *Config) CheckQuotaLimit() bool {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.usedQuota < c.Quota
}

// ReadConfig reads configuration from file
func ReadConfig(filepath string) (cfg *Config, err error) {
	var (
		data []byte
		e    error
	)
	if data, e = ioutil.ReadFile(filepath); err != nil {
		err = fmt.Errorf("Error reading file %s, %v", filepath, e)
		return
	}
	cfg = newConfig()
	if e = json.Unmarshal(data, &cfg); e != nil {
		err = fmt.Errorf("Error unmarshal config data from %s, %v", filepath, e)
		return
	}
	return
}
