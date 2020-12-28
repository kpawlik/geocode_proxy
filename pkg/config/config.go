package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultPort               = 8888
	defaultLogLevel           = "info"
	defaultWorkersNumber      = 10
	defaultQuota              = 0
	defaultQuotaTimeInMinutes = 0
)

// Geocoder part for google service
type Geocoder struct {
	APIKey       string `json:"apiKey"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Channel      string `json:"channel"`
}

// Server part of http configuration
type Server struct {
	Port     int    `json:"port"`
	LogLevel string `json:"logLevel"`
}

// Config stores configuration
type Config struct {
	Geocoder           Geocoder `json:"geocoder"`
	Server             Server   `json:"server"`
	WorkersNumber      int      `json:"workersNumber"`
	Quota              int      `json:"quota"`
	QuotaTimeInMinutes int      `json:"quotaTimeInMinutes"`
	usedQuotaCount     int
	useQuotaCheck      bool
	mux                sync.RWMutex
}

func newConfig() *Config {
	cfg := &Config{WorkersNumber: defaultWorkersNumber,
		mux:                sync.RWMutex{},
		Quota:              defaultQuota,
		QuotaTimeInMinutes: defaultQuotaTimeInMinutes,
		Server: Server{
			Port:     defaultPort,
			LogLevel: defaultLogLevel,
		},
	}
	return cfg
}

// SetUsedQuota set value of the used  quota
func (c *Config) SetUsedQuota(q int) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.usedQuotaCount = q
}

// IncQuota increments the value of used quota
func (c *Config) IncQuota() {
	c.mux.Lock()
	defer c.mux.Unlock()
	if !c.useQuotaCheck {
		return
	}
	c.usedQuotaCount++
}

// ResetUsedQuota set quota to default value
func (c *Config) ResetUsedQuota() {
	c.SetUsedQuota(0)
}

//GetRemainingQuota returns current quota to use
func (c *Config) GetRemainingQuota() int {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.Quota - c.usedQuotaCount
}

// IsAviableQuota checks if used quota exceeded
func (c *Config) IsAviableQuota() bool {
	c.mux.RLock()
	defer c.mux.RUnlock()
	if !c.useQuotaCheck {
		return true
	}
	return c.usedQuotaCount < c.Quota
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
	cfg.useQuotaCheck = cfg.Quota > 0
	return
}

// StartQuotaTimer start go function which checks if Quota should be resert
func StartQuotaTimer(cfg *Config) {
	if !cfg.useQuotaCheck {
		return
	}
	go func() {
		timeout := time.Duration(cfg.QuotaTimeInMinutes) * time.Minute
		timer := time.Tick(timeout)
		for range timer {
			log.Infof("Reset quota after timeout %v to value %d", timeout, cfg.Quota)
			cfg.ResetUsedQuota()
		}
	}()
}
