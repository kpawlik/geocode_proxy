package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kpawlik/geocode_server/pkg/config"
	"github.com/kpawlik/geocode_server/pkg/server"
	"github.com/sirupsen/logrus"
)

var (
	cfg    config.Config
	logger = logrus.New()
)

func init() {
	var (
		configPath string
		err        error
	)

	flag.StringVar(&configPath, "config", "config.json", "Path to config file")
	flag.Parse()
	if configPath == "" {
		fmt.Printf("Missing config file path")
		flag.Usage()
		os.Exit(1)
	}
	if cfg, err = config.ReadConfig(configPath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	setLogger(cfg)
	server.Serve(cfg, logger)
}

func setLogger(cfg config.Config) {
	switch cfg.LogLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
		break
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
		break
	default:
		logger.SetLevel(logrus.InfoLevel)
		break
	}
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: false,
	})
	logger.SetOutput(os.Stdout)
}
