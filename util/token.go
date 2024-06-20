package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

const tokenPrefix = "qrystalct_"
const hashPrefix = "qrystalcth_"

type TokenHash struct {
	raw [32]byte
}

func (h *TokenHash) String() string {
	return hashPrefix + hex.EncodeToString(h.raw[:])
}

func ParseTokenHash(s string) (*TokenHash, error) {
	raw, err := parseTokenHash(s)
	if err != nil {
		return nil, err
	}
	return &TokenHash{raw: raw}, nil
}

func parseTokenHash(s string) ([32]byte, error) {
	if !strings.HasPrefix(s, hashPrefix) {
		return [32]byte{}, errors.New("missing prefix")
	}
	var r [32]byte
	_, err := hex.Decode(r[:], []byte(s[len(hashPrefix):]))
	if err != nil {
		return [32]byte{}, err
	}
	return r, nil
}

// MarshalJSON implements json.Marshaler.
func (h *TokenHash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (h *TokenHash) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	raw, err := parseTokenHash(s)
	if err != nil {
		return err
	}
	h.raw = raw
	return nil
}

// GobEncode implements gob.GobDecoder.
func (h *TokenHash) GobEncode() ([]byte, error) {
	return []byte(h.String()), nil
}

// GobDecode implements gob.GobDecoder.
func (h *TokenHash) GobDecode(data []byte) error {
	raw, err := parseTokenHash(string(data))
	if err != nil {
		return err
	}
	h.raw = raw
	return nil
}

type Token struct {
	raw []byte
}

func RandomToken() (*Token, error) {
	raw := make([]byte, 64)
	_, err := rand.Read(raw)
	if err != nil {
		return nil, err
	}
	return &Token{raw: raw}, nil
}

func ParseToken(s string) (*Token, error) {
	raw, err := parseToken(s)
	if err != nil {
		return nil, err
	}
	return &Token{raw: raw}, nil
}

func parseToken(s string) ([]byte, error) {
	if !strings.HasPrefix(s, tokenPrefix) {
		return nil, errors.New("missing prefix")
	}
	raw, err := base64.StdEncoding.DecodeString(s[len(tokenPrefix):])
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (t *Token) Empty() bool {
	return len(t.raw) == 0
}

func (t *Token) Hash() *TokenHash {
	sum := sha256.Sum256(t.raw)
	return &TokenHash{raw: sum}
}

func (t *Token) String() string {
	return tokenPrefix + base64.StdEncoding.EncodeToString(t.raw)
}

// MarshalJSON implements json.Marshaler.
func (t *Token) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Token) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	raw, err := parseToken(s)
	if err != nil {
		return err
	}
	t.raw = raw
	return nil
}

// GobEncode implements gob.GobDecoder.
func (t *Token) GobEncode() ([]byte, error) {
	return []byte(t.String()), nil
}

// GobDecode implements gob.GobDecoder.
func (t *Token) GobDecode(data []byte) error {
	raw, err := parseToken(string(data))
	if err != nil {
		return err
	}
	t.raw = raw
	return nil
}
