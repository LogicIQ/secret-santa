package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

type ChaCha20KeyGenerator struct{}

func generateChaCha20Key(algorithm string) (map[string]string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return map[string]string{
		"key_base64": base64.StdEncoding.EncodeToString(key),
		"key_hex":    hex.EncodeToString(key),
		"key_size":   "256",
		"algorithm":  algorithm,
	}, nil
}

func (g *ChaCha20KeyGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	// config is unused but required by Generator interface
	return generateChaCha20Key("ChaCha20")
}

type XChaCha20KeyGenerator struct{}

func (g *XChaCha20KeyGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	// config is unused but required by Generator interface
	return generateChaCha20Key("XChaCha20")
}
