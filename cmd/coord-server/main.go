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
	Spec   spec.Spec
	Tokens map[string]coord.TokenInfo
	Addr   string
}

func main() {
	var configPath string
	var addr string
	flag.StringVar(&configPath, "config", "", "config file path")
	flag.StringVar(&addr, "addr", "", "bind address")
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
	tokens, err := convertTokens(c.Tokens)
	if err != nil {
		zap.S().Fatalf("loading config failed: %s", err)
	}
	s := coord.NewServer(c.Spec, tokens)
	err = util.Notify("READY=1\nSTATUS=serving…")
	if err != nil {
		zap.S().Infof("notify: %s", err)
	}
	zap.S().Fatalf("listen and serve failed: %s", http.ListenAndServe(addr, s))
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
