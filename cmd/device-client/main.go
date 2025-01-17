package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"github.com/nyiyui/qrystal/device"
	"github.com/nyiyui/qrystal/dns"
	"github.com/nyiyui/qrystal/goal"
	"github.com/nyiyui/qrystal/util"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Config struct {
	Clients    map[string]ClientConfig
	CanForward bool
	AssumeProc bool
}

type ClientConfig struct {
	BaseURL         string
	Token           util.Token
	TokenPath       string
	Network         string
	Device          string
	PrivateKey      goal.Key
	PrivateKeyPath  string
	MinimumInterval goal.Duration
	CertPath        string
	transport       *http.Transport
}

func main() {
	util.SetupLog()

	var configPath string
	var dnsSocketPath string
	var dnsConfigPath string
	var dnsAddr string
	var dnsSelf bool
	flag.StringVar(&configPath, "config", "", "path to config file (required)")
	flag.StringVar(&dnsSocketPath, "dns-socket", "", "socket to connect to DNS server (optional)")
	flag.StringVar(&dnsConfigPath, "dns-config", "", "path to DNS config file (required for -dns-self)")
	flag.StringVar(&dnsAddr, "dns-addr", "", "address to listen on for DNS")
	flag.BoolVar(&dnsSelf, "dns-self", false, "act as the DNS server itself")
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
		if cc.PrivateKeyPath != "" {
			data, err := os.ReadFile(cc.PrivateKeyPath)
			if err != nil {
				zap.S().Fatalf("parsing config file failed: client %s: reading private key path: %s", key, err)
			}
			privateKey, err := wgtypes.ParseKey(string(data))
			if err != nil {
				zap.S().Fatalf("parsing config file failed: client %s: parsing private key path: %s", key, err)
			}
			cc.PrivateKey = goal.Key(privateKey)
		}
		if cc.TokenPath != "" {
			data, err := os.ReadFile(cc.TokenPath)
			if err != nil {
				zap.S().Fatalf("parsing config file failed: client %s: reading token path: %s", key, err)
			}
			tok, err := util.ParseToken(string(data))
			if err != nil {
				zap.S().Fatalf("parsing config file failed: client %s: parsing token path: %s", key, err)
			}
			cc.Token = *tok
		}
		if cc.CertPath != "" {
			pool := x509.NewCertPool()
			cert, err := os.ReadFile(cc.CertPath)
			if err != nil {
				zap.S().Fatalf("parsing config file failed: client %s: reading cert file: %s", key, err)
			}
			ok := pool.AppendCertsFromPEM(cert)
			if !ok {
				zap.S().Fatalf("parsing config file failed: client %s: appending cert failed", key)
			}
			zap.S().Infof("client %s: loaded cert from %s", key, cc.CertPath)
			cc.transport = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}
		}
		config.Clients[key] = cc
	}
	data, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	zap.S().Infof("parsed config:\n%s", data)

	var dnsClient dns.Client
	if dnsSocketPath != "" {
		zap.S().Infof("connecting to DNS server at %s…", dnsSocketPath)
		rpcClient, err := rpc.Dial("unix", dnsSocketPath)
		if err != nil {
			zap.S().Fatalf("connecting to DNS server failed: %s", err)
		}
		dnsClient = dns.NewRPCClient(rpcClient)
		zap.S().Info("done connecting to DNS server.")
	} else if dnsSelf {
		configData, err := os.ReadFile(dnsConfigPath)
		if err != nil {
			zap.S().Fatalf("reading DNS config file failed: %s", err)
		}
		var config dns.Config
		err = json.Unmarshal(configData, &config)
		if err != nil {
			zap.S().Fatalf("parsing DNS config file failed: %s", err)
		}
		if dnsAddr == "" {
			dnsAddr = config.Address
		}
		data, err := json.Marshal(config)
		if err != nil {
			panic(err)
		}
		zap.S().Infof("parsed DNS config:\n%s", data)

		s, err := dns.NewServer(config.Parents)
		if err != nil {
			zap.S().Fatalf("%s", err)
		}
		err = s.ListenDNS(dnsAddr)
		if err != nil {
			zap.S().Fatalf("failed to listen: %s", err)
		}
		zap.S().Infof("listening for DNS on %s.", dnsAddr)
		dnsClient = dns.NewDirectClient(s)
	}

	createGoroutines(dnsClient, config)
}

func createGoroutines(dnsClient dns.Client, config Config) {
	util.Notify("READY=1\nSTATUS=starting…")
	for clientName, cc := range config.Clients {
		go func(clientName string, cc ClientConfig) {
			c, err := device.NewClient(&http.Client{
				Timeout: 5 * time.Second,
				//Transport: cc.transport, // TODO: nil pointer
			}, cc.BaseURL, cc.Token, cc.Network, cc.Device, cc.PrivateKey)
			if err != nil {
				panic(err)
			}
			c.SetCanForward(config.CanForward)
			c.SetAssumeProc(config.AssumeProc)
			c.SetDNSClient(dnsClient)
			continuous := new(device.ContinousClient)
			continuous.Client = c
			zap.S().Infof("%s: created client.", clientName)

			t := time.NewTicker(time.Duration(cc.MinimumInterval))
			for {
				latest := false
				for !latest {
					var updated bool
					var err error
					latest, updated, err = continuous.Step()
					if err != nil {
						zap.S().Errorf("%s: %s", clientName, err)
						break
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
				<-t.C
			}
		}(clientName, cc)
	}
	select {}
}
