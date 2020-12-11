package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kpawlik/geocode_server/pkg/config"
	"github.com/kpawlik/geocode_server/pkg/server"
)

var (
	cfg config.Config
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

	server.Serve(cfg)
}
