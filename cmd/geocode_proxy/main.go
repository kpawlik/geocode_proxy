package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kpawlik/geocode_proxy/pkg/config"
	"github.com/kpawlik/geocode_proxy/pkg/server"
	log "github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
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
	if level, ok := logLevelMap[cfg.Log.LogLevel]; ok {
		logLevel = level
	} else {
		logLevel = defaultLogLevel
	}
	log.SetLevel(logLevel)
	if cfg.Log.Format == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: false,
			FullTimestamp: false,
		})
	}
	if cfg.Log.Stdout {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(&lumberjack.Logger{
			Filename:   filepath.Join(cfg.Log.Directory, cfg.Log.Filename),
			MaxSize:    10, // megabytes
			MaxBackups: 10,
		})
	}

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
