package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nyiyui/qrystal/device"
	"github.com/nyiyui/qrystal/goal"
	"github.com/nyiyui/qrystal/util"
	"go.uber.org/zap"
)

type Config struct {
	Clients map[string]ClientConfig
}

type ClientConfig struct {
	BaseURL         string
	Token           util.Token
	Network         string
	Device          string
	PrivateKey      goal.Key
	MinimumInterval time.Duration
}

type MachineData struct {
	Machines     map[string]goal.Machine
	machinesLock sync.RWMutex
	path         string
}

func LoadMachineData(path string) (*MachineData, error) {
	var m MachineData
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading failed: %w", err)
	}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}
	m.path = path
	return &m, nil
}

func (m *MachineData) setMachine(clientName string, gm goal.Machine) {
	m.machinesLock.Lock()
	defer m.machinesLock.Unlock()
	m.Machines[clientName] = gm
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(m.path, data, 0600)
	if err != nil {
		zap.S().Errorf("saving machine data failed: %s. Updating WireGuard interfaces may fail.", err)
	}
}

func main() {
	util.SetupLog()

	var configPath string
	flag.StringVar(&configPath, "config-path", "", "path to config file")
	flag.Parse()
	configData, err := os.ReadFile(configPath)
	if err != nil {
		zap.S().Fatalf("reading config file failed: %s", err)
	}
	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		zap.S().Fatalf("parsing config file failed: %s", err)
	}
	data, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	zap.S().Infof("parsed config:\n%s", data)

	m, err := LoadMachineData(filepath.Join(os.Getenv("STATE_DIRECTORY"), "MachineData.json"))
	if err != nil {
		zap.S().Fatalf("loading machine data failed: %s", err)
	}
	createGoroutines(m, config)
}

func createGoroutines(m *MachineData, config Config) {
	for clientName, clientConfig := range config.Clients {
		go func(clientName string, clientConfig ClientConfig) {
			c, err := device.NewClient(clientConfig.BaseURL, clientConfig.Token, clientConfig.Network, clientConfig.Device, clientConfig.PrivateKey)
			if err != nil {
				panic(err)
			}
			cc := new(device.ContinousClient)
			cc.Client = c
			zap.S().Infof("%s: created client.", clientName)

			t := time.NewTicker(clientConfig.MinimumInterval)
			for range t.C {
				latest := false
				for !latest {
					var updated bool
					var err error
					updated, latest, err = cc.Step()
					if err != nil {
						zap.S().Errorf("%s: %s", clientName, err)
					}
					if updated {
						m.setMachine(clientName, c.Machine)
					}
					if !latest {
						if updated {
							zap.S().Infof("%s: updated but not latest; trying again…", clientName)
						} else {
							zap.S().Infof("%s: not latest anymore; trying to update…", clientName)
						}
					} else {
						if updated {
							zap.S().Infof("%s: updated.", clientName)
						} else {
							zap.S().Infof("%s: latest.", clientName)
						}
						break
					}
				}
			}
		}(clientName, clientConfig)
	}
}
