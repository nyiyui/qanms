package cs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/tidwall/buntdb"
)

const tokenPrefix = "token-"

type sha256Sum = [sha256.Size]byte

type tokenStore struct {
	db *buntdb.DB
}

func newTokenStore(db *buntdb.DB) (tokenStore, error) {
	err := db.CreateIndex("tokens", "token:*", buntdb.IndexString)
	return tokenStore{
		db: db,
	}, err
}

func (s *tokenStore) AddToken(sum sha256Sum, info TokenInfo, overwrite bool) (alreadyExists bool, err error) {
	encoded, err := json.Marshal(info)
	if err != nil {
		return
	}
	key := tokenPrefix + hex.EncodeToString(sum[:])
	err = s.db.Update(func(tx *buntdb.Tx) error {
		_, err = tx.Get(key)
		if err == nil && !overwrite {
			alreadyExists = true
			return nil
		}
		if err != buntdb.ErrNotFound {
			return err
		}
		_, _, err = tx.Set(key, string(encoded), nil)
		alreadyExists = false
		return nil
	})
	return
}

func (s *tokenStore) getToken(token string) (info TokenInfo, ok bool, err error) {
	sum := sha256.Sum256([]byte(token))
	key := tokenPrefix + hex.EncodeToString(sum[:])
	var encoded string
	err = s.db.View(func(tx *buntdb.Tx) error {
		encoded, err = tx.Get(key)
		return err
	})
	if err == buntdb.ErrNotFound {
		ok = false
		return
	}
	err = json.Unmarshal([]byte(encoded), &info)
	ok = true
	return
}

func (s *tokenStore) convertToMap() (m map[string]string, err error) {
	m = map[string]string{}
	err = s.db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend("tokens", func(key, val string) bool {
			m[key] = val
			return true
		})
	})
	return
}

type TokenInfo struct {
	Name         string
	Networks     map[string]string
	CanPull      bool
	CanPush      *CanPush
	CanAddTokens *CanAddTokens
}

type CanAddTokens struct {
	CanPull bool `yaml:"can-pull"`
	CanPush bool `yaml:"can-push"`
	// don't allow CanAddTokens to make logic simpler
}

type CanPush struct {
	Any      bool              `yaml:"any"`
	Networks map[string]string `yaml:"networks"`
}

type Token struct {
	Hash [sha256.Size]byte
	Info TokenInfo
}

func convertTokens(tokens []Token) map[[sha256.Size]byte]TokenInfo {
	m := map[[sha256.Size]byte]TokenInfo{}
	for _, token := range tokens {
		m[token.Hash] = token.Info
	}
	return m
}

func newTokenAuthError(token string) error {
	sum := sha256.Sum256([]byte(token))
	return fmt.Errorf("token auth failed with hash %x", sum)
}
