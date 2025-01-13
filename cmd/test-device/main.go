package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/nyiyui/qrystal/device"
	"github.com/nyiyui/qrystal/goal"
	"github.com/nyiyui/qrystal/util"
	"go.uber.org/zap"
)

type Config struct {
	BaseURL    string
	Token      util.Token
	Network    string
	Device     string
	PrivateKey goal.Key
}

func main() {
	util.SetupLog()

	var configPath string
	flag.StringVar(&configPath, "config-path", "", "path to config file")
	flag.Parse()
	configData, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		panic(err)
	}
	zap.S().Info("parsed configuration.")

	c, err := device.NewClient(nil, config.BaseURL, config.Token, config.Network, config.Device, config.PrivateKey)
	if err != nil {
		panic(err)
	}
	zap.S().Info("created client.")

	latest, err := c.ReifySpec()
	if err != nil {
		panic(err)
	}
	if !latest {
		panic("posted but already not latest")
	}
	zap.S().Info("reified spec.")
}
