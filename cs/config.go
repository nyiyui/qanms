package cs

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/nyiyui/qrystal/node"
	"github.com/nyiyui/qrystal/util"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Addr        string              `yaml:"addr"`
	TLSCertPath string              `yaml:"tls-cert-path"`
	TLSKeyPath  string              `yaml:"tls-key-path"`
	CC          *node.CentralConfig `yaml:"central"`
	Tokens      *TokensConfig       `yaml:"tokens"`
}

type TokensConfig struct {
	Raw []Token
}

func (t *TokensConfig) UnmarshalYAML(value *yaml.Node) error {
	var raw []TokenConfig
	err := value.Decode(&raw)
	if err != nil {
		return err
	}
	t2, err := convertTokens2(raw)
	if err != nil {
		return err
	}
	t.Raw = t2
	return nil
}

type TokenConfig struct {
	Name         string            `yaml:"name"`
	Hash         *util.HexBytes    `yaml:"hash"`
	Networks     map[string]string `yaml:"networks"`
	CanPull      bool              `yaml:"can-pull"`
	CanPush      bool              `yaml:"can-push"`
	CanAddTokens *CanAddTokens     `yaml:"can-add-tokens"`
}

func convertTokens2(tokens []TokenConfig) ([]Token, error) {
	res := make([]Token, len(tokens))
	for i, token := range tokens {
		var hash [sha256.Size]byte
		log.Println(len(hash))
		n := copy(hash[:], *token.Hash)
		if n != len(hash) {
			return nil, fmt.Errorf("token %d: invalid length (%d) hash", i, n)
		}
		res[i] = Token{
			Hash: hash,
			Info: TokenInfo{
				Name:         token.Name,
				Networks:     token.Networks,
				CanPull:      token.CanPull,
				CanPush:      token.CanPush,
				CanAddTokens: token.CanAddTokens,
			},
		}
	}
	return res, nil
}

func LoadConfig(configPath string) (*Config, error) {
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config read: %s", err)
	}
	var config Config
	err = yaml.Unmarshal(raw, &config)
	if err != nil {
		return nil, fmt.Errorf("config unmarshal: %s", err)
	}
	for cnn, cn := range config.CC.Networks {
		if cn.Me != "" {
			return nil, fmt.Errorf("net %s: me is not blank", cnn)
		}
	}
	return &config, nil
}