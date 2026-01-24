package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

type StringGenerator struct{}

func (g *StringGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	length := getIntConfig(config, "length", 16)
	if length <= 0 {
		return nil, fmt.Errorf("string length must be positive, got %d", length)
	}
	if length > 10000 {
		return nil, fmt.Errorf("string length too large, maximum 10000, got %d", length)
	}
	lower := getBoolConfig(config, "lower", true)
	upper := getBoolConfig(config, "upper", true)
	numeric := getBoolConfig(config, "numeric", true)
	special := getBoolConfig(config, "special", true)
	overrideSpecial := getStringConfig(config, "override_special", "")

	// Build character set
	var charset strings.Builder
	if lower {
		charset.WriteString("abcdefghijklmnopqrstuvwxyz")
	}
	if upper {
		charset.WriteString("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	}
	if numeric {
		charset.WriteString("0123456789")
	}
	if special {
		if overrideSpecial != "" {
			charset.WriteString(overrideSpecial)
		} else {
			charset.WriteString("!@#$%&*()-_=+[]{}<>:?")
		}
	}

	charsetStr := charset.String()
	if len(charsetStr) == 0 {
		return nil, fmt.Errorf("no character types enabled")
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charsetStr)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random character: %w", err)
		}
		result[i] = charsetStr[n.Int64()]
	}

	return map[string]string{
		"value":   string(result),
		"charset": charsetStr,
	}, nil
}
