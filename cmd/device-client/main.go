package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net/rpc"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nyiyui/qrystal/device"
	"github.com/nyiyui/qrystal/dns"
	"github.com/nyiyui/qrystal/goal"
	"github.com/nyiyui/qrystal/util"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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
	PrivateKeyPath  string
	MinimumInterval goal.Duration
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
	var dnsSocketPath string
	flag.StringVar(&configPath, "config", "", "path to config file (required)")
	flag.StringVar(&dnsSocketPath, "dns-socket", "", "socket to connect to DNS server (optional)")
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
	for key, cc := range config.Clients {
		if cc.PrivateKeyPath == "" {
			continue
		}
		data, err := os.ReadFile(cc.PrivateKeyPath)
		if err != nil {
			zap.S().Fatalf("parsing config file failed: client %s: reading private key path: %s", key, err)
		}
		privateKey, err := wgtypes.ParseKey(string(data))
		if err != nil {
			zap.S().Fatalf("parsing config file failed: client %s: parsing private key path: %s", key, err)
		}
		cc.PrivateKey = goal.Key(privateKey)
		config.Clients[key] = cc
	}
	data, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	zap.S().Infof("parsed config:\n%s", data)

	var dnsClient *dns.RPCClient
	if dnsSocketPath != "" {
		zap.S().Infof("connecting to DNS server at %s…", dnsSocketPath)
		rpcClient, err := rpc.Dial("unix", dnsSocketPath)
		if err != nil {
			zap.S().Fatalf("connecting to DNS server failed: %s", err)
		}
		dnsClient = dns.NewRPCClient(rpcClient)
		zap.S().Info("done connecting to DNS server.")
	}

	path := filepath.Join(os.Getenv("STATE_DIRECTORY"), "MachineData.json")
	m, err := LoadMachineData(path)
	if err != nil {
		zap.S().Debugf("IsNotExist=%t", os.IsNotExist(err))
		if errors.Is(err, fs.ErrNotExist) {
			zap.S().Info("no machine data found. Creating a blank one…")
			m = new(MachineData)
			m.Machines = map[string]goal.Machine{}
			m.path = path
		} else {
			zap.S().Fatalf("loading machine data failed: %s", err)
		}
	}
	createGoroutines(m, dnsClient, config)
}

func createGoroutines(m *MachineData, dnsClient *dns.RPCClient, config Config) {
	util.Notify("READY=1\nSTATUS=starting…")
	for clientName, clientConfig := range config.Clients {
		go func(clientName string, clientConfig ClientConfig) {
			c, err := device.NewClient(clientConfig.BaseURL, clientConfig.Token, clientConfig.Network, clientConfig.Device, clientConfig.PrivateKey)
			if err != nil {
				panic(err)
			}
			c.SetDNSClient(dnsClient)
			cc := new(device.ContinousClient)
			cc.Client = c
			zap.S().Infof("%s: created client.", clientName)

			t := time.NewTicker(time.Duration(clientConfig.MinimumInterval))
			for range t.C {
				latest := false
				for !latest {
					var updated bool
					var err error
					latest, updated, err = cc.Step()
					if err != nil {
						zap.S().Errorf("%s: %s", clientName, err)
						break
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
					zap.S().Info("sleeping 1 second until next loop.")
					time.Sleep(1 * time.Second)
				}
			}
		}(clientName, clientConfig)
	}
	select {}
}
