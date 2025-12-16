package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

type ChaCha20KeyGenerator struct{}

func (g *ChaCha20KeyGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	// ChaCha20 uses 32-byte keys
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return map[string]string{
		"key_base64": base64.StdEncoding.EncodeToString(key),
		"key_hex":    hex.EncodeToString(key),
		"key_size":   "256",
		"algorithm":  "ChaCha20",
	}, nil
}

type XChaCha20KeyGenerator struct{}

func (g *XChaCha20KeyGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	// XChaCha20 uses 32-byte keys
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return map[string]string{
		"key_base64": base64.StdEncoding.EncodeToString(key),
		"key_hex":    hex.EncodeToString(key),
		"key_size":   "256",
		"algorithm":  "XChaCha20",
	}, nil
}
