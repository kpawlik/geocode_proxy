package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kpawlik/geocode_server/pkg/config"
	"github.com/kpawlik/geocode_server/pkg/server"
	log "github.com/sirupsen/logrus"
)

var (
	cfg *config.Config
	// logger = logrus.New()
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
	server.Serve(cfg)
}

func setLogger(cfg *config.Config) {
	switch cfg.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
		break
	case "warn":
		log.SetLevel(log.WarnLevel)
		break
	case "error":
		log.SetLevel(log.ErrorLevel)
		break
	default:
		log.SetLevel(log.InfoLevel)
		break
	}
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: false,
	})
	log.SetOutput(os.Stdout)
}
