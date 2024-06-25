package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/coreos/go-systemd/v22/activation"

	"github.com/nyiyui/qrystal/dns"
	"github.com/nyiyui/qrystal/util"
	"go.uber.org/zap"
)

type Config struct {
	Parents []dns.Parent
	Address string
}

func main() {
	util.SetupLog()

	var configPath string
	var socketPath string
	var useSystemdSocketActivation bool
	var addr string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&socketPath, "rpc-listen", "", "socket to listen on for RPC. NOTE that sockets must be made in a private parent directory, as anyone with access to this socket has access to a DNS server running as root.")
	flag.BoolVar(&useSystemdSocketActivation, "rpc-systemd", false, "use systemd socket activation for RPC listening.")
	flag.StringVar(&addr, "dns-listen", "", "socket to listen on for DNS")
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
	if addr == "" {
		addr = config.Address
	}
	data, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	zap.S().Infof("parsed config:\n%s", data)

	s, err := dns.NewServer(config.Parents)
	if err != nil {
		zap.S().Fatalf("%s", err)
	}
	if useSystemdSocketActivation {
		listeners, err := activation.Listeners()
		if err != nil {
			zap.S().Fatalf("getting socket activation listeners failed: %s", err)
		}
		if len(listeners) != 1 {
			zap.S().Fatalf("Unexpected number of socket activation fds (got %d, want 1)", len(listeners))
		}
		s.ListenRPCListener(listeners[0])
		zap.S().Info("listening for RPC on socket activation.")
	} else {
		err = s.ListenRPC(socketPath)
		if err != nil {
			zap.S().Fatalf("failed to listen: %s", err)
		}
		zap.S().Infof("listening for RPC on %s.", socketPath)
	}
	err = s.ListenDNS(addr)
	if err != nil {
		zap.S().Fatalf("failed to listen: %s", err)
	}
	zap.S().Infof("listening for DNS on %s.", addr)
	util.Notify("READY=1\nSTATUS=listening on both RPC and DNS…")
	select {}
}
