package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// AESKeyGenerator generates AES encryption keys
type AESKeyGenerator struct{}

func (g *AESKeyGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	keySize := getIntConfig(config, "key_size", 256)

	// Validate key size
	var keyBytes int
	switch keySize {
	case 128:
		keyBytes = 16
	case 192:
		keyBytes = 24
	case 256:
		keyBytes = 32
	default:
		return nil, fmt.Errorf("invalid key_size: %d (must be 128, 192, or 256)", keySize)
	}

	// Generate random key
	key := make([]byte, keyBytes)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return map[string]string{
		"key_base64": base64.StdEncoding.EncodeToString(key),
		"key_hex":    hex.EncodeToString(key),
		"key_size":   fmt.Sprintf("%d", keySize),
	}, nil
}
