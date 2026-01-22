package random

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

type BytesGenerator struct{}

func (g *BytesGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	length := getIntConfig(config, "length", 16)

	if length < 1 {
		return nil, fmt.Errorf("length must be at least 1")
	}

	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"value":       base64.StdEncoding.EncodeToString(buf),
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}
