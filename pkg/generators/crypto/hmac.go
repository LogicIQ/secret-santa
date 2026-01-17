package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"
)

type HMACGenerator struct{}

func (g *HMACGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	algorithm := strings.ToLower(strings.TrimSpace(getStringConfig(config, "algorithm", "sha256")))
	keySize := getIntConfig(config, "key_size", 32)
	message := getStringConfig(config, "message", "")

	// Validate and sanitize algorithm early to prevent log injection
	if algorithm == "" {
		return nil, fmt.Errorf("algorithm cannot be empty")
	}
	// Sanitize algorithm to prevent log injection
	algorithm = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, algorithm)
	if algorithm == "" {
		return nil, fmt.Errorf("algorithm contains invalid characters")
	}

	if keySize <= 0 {
		return nil, fmt.Errorf("key_size must be positive, got: %d", keySize)
	}
	if keySize > 1024 {
		return nil, fmt.Errorf("key_size too large, maximum 1024 bytes, got: %d", keySize)
	}

	// Generate random key if not provided
	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	// Select hash function
	var hashFunc func() hash.Hash
	switch algorithm {
	case "sha256":
		hashFunc = sha256.New
	case "sha512":
		hashFunc = sha512.New
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s (supported: sha256, sha512)", algorithm)
	}

	// Generate HMAC
	h := hmac.New(hashFunc, key)
	h.Write([]byte(message))
	signature := h.Sum(nil)

	return map[string]string{
		"key_base64":       base64.StdEncoding.EncodeToString(key),
		"key_hex":          hex.EncodeToString(key),
		"signature_base64": base64.StdEncoding.EncodeToString(signature),
		"signature_hex":    hex.EncodeToString(signature),
		"algorithm":        algorithm,
	}, nil
}
