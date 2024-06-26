package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/nyiyui/qrystal/coord"
	"github.com/nyiyui/qrystal/profile"
	"github.com/nyiyui/qrystal/spec"
	"github.com/nyiyui/qrystal/util"
	"go.uber.org/zap"
)

type Config struct {
	Spec     spec.Spec
	Tokens   map[string]coord.TokenInfo
	Addr     string
	CertPath string
	KeyPath  string
}

func main() {
	var configPath string
	var addr string
	var certPath string
	var keyPath string
	flag.StringVar(&configPath, "config", "", "Config file path.")
	flag.StringVar(&addr, "addr", "", "Bind address. Overridden by config file if present.")
	flag.StringVar(&certPath, "cert", "", "Certificate for HTTPS server. Supplying this will enable HTTPS and disable HTTP. Overridden by config file if present.")
	flag.StringVar(&keyPath, "key", "", "Key for HTTPS server. Supplying this will enable HTTPS and disable HTTP. Overridden by config file if present.")
	flag.Parse()
	util.SetupLog()
	defer util.S.Sync()
	profile.Profile()

	c, err := loadConfig(configPath)
	if err != nil {
		zap.S().Fatalf("loading config failed: %s", err)
	}
	if c.Addr != "" {
		addr = c.Addr
	}
	if c.CertPath != "" {
		certPath = c.CertPath
	}
	if c.KeyPath != "" {
		keyPath = c.KeyPath
	}
	if (certPath == "") != (keyPath == "") {
		zap.S().Fatalf("both or none of certPath and keyPath must be provided")
	}
	tokens, err := convertTokens(c.Tokens)
	if err != nil {
		zap.S().Fatalf("loading config failed: %s", err)
	}
	s := coord.NewServer(c.Spec, tokens)
	if certPath != "" && keyPath != "" {
		err = util.Notify("READY=1\nSTATUS=serving HTTPS…")
	} else {
		err = util.Notify("READY=1\nSTATUS=serving HTTP…")
	}
	if err != nil {
		zap.S().Infof("notify: %s", err)
	}
	if certPath != "" && keyPath != "" {
		zap.S().Fatalf("listen and serve failed: %s", http.ListenAndServeTLS(addr, certPath, keyPath, s))
	} else {
		zap.S().Fatalf("listen and serve failed: %s", http.ListenAndServe(addr, s))
	}
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var c Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		return Config{}, err
	}
	return c, nil
}

func convertTokens(tokens map[string]coord.TokenInfo) (map[util.TokenHash]coord.TokenInfo, error) {
	tokens2 := map[util.TokenHash]coord.TokenInfo{}
	for key, val := range tokens {
		tokenHash, err := util.ParseTokenHash(key)
		if err != nil {
			return nil, fmt.Errorf("parsing token hash %s: %w", key, err)
		}
		tokens2[*tokenHash] = val
	}
	return tokens2, nil
}
