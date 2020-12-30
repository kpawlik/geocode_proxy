package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kpawlik/geocode_server/pkg/config"
	"github.com/kpawlik/geocode_server/pkg/server"
	log "github.com/sirupsen/logrus"
)

const version = "0.1.202012301447"

var (
	configPath  string
	logLevelMap = map[string]log.Level{
		"debug":   log.DebugLevel,
		"warning": log.WarnLevel,
		"error":   log.ErrorLevel,
		"info":    log.InfoLevel,
	}
	defaultLogLevel log.Level = log.InfoLevel
)

func init() {
	var (
		printVersion bool
	)
	flag.StringVar(&configPath, "config", "config.json", "Path to config file")
	flag.BoolVar(&printVersion, "version", false, "Print version")
	flag.Parse()
	if printVersion {
		fmt.Printf("Version: %s", version)
		os.Exit(0)
	}
}

func setLogger(cfg *config.Config) {
	var (
		logLevel log.Level
	)
	if level, ok := logLevelMap[cfg.Server.LogLevel]; ok {
		logLevel = level
	} else {
		logLevel = defaultLogLevel
	}
	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})
	log.SetOutput(os.Stdout)
}

func main() {
	var (
		err error
		cfg *config.Config
	)
	if cfg, err = config.ReadConfig(configPath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	setLogger(cfg)
	log.Debugf("Config: %+v", cfg)
	log.Info("Server start")
	if err := server.Serve(cfg); err != nil {
		log.Error(err)
	}
}
