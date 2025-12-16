package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type StringGenerator struct{}

func (g *StringGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	length := getIntConfig(config, "length", 16)
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
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charsetStr))))
		if err != nil {
			return nil, err
		}
		result[i] = charsetStr[n.Int64()]
	}

	return map[string]string{
		"value":       string(result),
		"charset":     charsetStr,
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}
