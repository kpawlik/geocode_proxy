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
	defaultPort          = 9998
	defaultLogLevel      = "info"
	defaultLogFormat     = "text"
	defaultWorkersNumber = 10
	defaultQuota         = 0 //unlimited
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
	Port int `json:"port"`
}

// Log configuration
type Log struct {
	LogLevel  string `json:"logLevel"`
	Stdout    bool   `json:"stdout"`
	Format    string `json:"format"`
	Filename  string `json:"filename"`
	Directory string `json:"diretory"`
}

// Config stores configuration
type Config struct {
	Geocoder       Geocoder `json:"geocoder"`
	Server         Server   `json:"server"`
	Log            Log      `json:"log"`
	WorkersNumber  int      `json:"workersNumber"`
	Quota          int      `json:"quota"`
	QuotaTime      string   `json:"quotaTime"`
	quotaTime      time.Duration
	usedQuotaCount int
	useQuotaCheck  bool
	mux            sync.RWMutex
}

func newConfig() *Config {
	cfg := &Config{WorkersNumber: defaultWorkersNumber,
		mux:   sync.RWMutex{},
		Quota: defaultQuota,
		Server: Server{
			Port: defaultPort,
		},
		Log: Log{
			LogLevel: defaultLogLevel,
			Stdout:   false,
			Format:   defaultLogFormat,
			Filename: "geocode_proxy_server.log",
		},
	}
	return cfg
}

func (c *Config) setUsedQuota(q int) {
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
	c.setUsedQuota(0)
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
	if len(cfg.QuotaTime) != 0 {
		var qt time.Duration
		if qt, err = time.ParseDuration(cfg.QuotaTime); err != nil {
			err = fmt.Errorf("Config error: Wrong value in quotaTime")
			return
		}
		cfg.quotaTime = qt

	}
	return
}

// StartQuotaTimer start go function which checks if Quota should be resert
func StartQuotaTimer(cfg *Config) {
	if !cfg.useQuotaCheck || cfg.quotaTime == 0 {
		return
	}
	go func() {
		timeout := time.Duration(cfg.quotaTime)
		timer := time.Tick(timeout)
		for range timer {
			log.Infof("Reset quota after timeout %v to value %d", timeout, cfg.Quota)
			cfg.ResetUsedQuota()
		}
	}()
}
