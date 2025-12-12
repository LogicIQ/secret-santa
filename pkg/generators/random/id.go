package random

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// IDGenerator generates random IDs
type IDGenerator struct{}

func (g *IDGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	byteLength := getIntConfig(config, "byte_length", 8)
	prefix := getStringConfig(config, "prefix", "")

	if byteLength < 1 {
		return nil, fmt.Errorf("byte_length must be at least 1")
	}

	bytes := make([]byte, byteLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}

	hexStr := hex.EncodeToString(bytes)

	value := hexStr
	if prefix != "" {
		value = prefix + hexStr
	}

	return map[string]string{
		"value":       value,
		"prefix":      prefix,
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}
