package random

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type BytesGenerator struct{}

func (g *BytesGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	length := getIntConfig(config, "length", 16)

	if length < 1 {
		return nil, fmt.Errorf("length must be at least 1")
	}
	if length > 1024 {
		return nil, fmt.Errorf("length too large, maximum 1024")
	}

	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"value": base64.StdEncoding.EncodeToString(buf),
	}, nil
}
