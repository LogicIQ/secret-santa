package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type PasswordGenerator struct{}

func (g *PasswordGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	length := getIntConfig(config, "length", 16)
	if length <= 0 || length > 1000000 {
		return nil, fmt.Errorf("password length must be between 1 and 1000000, got %d", length)
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

	charsetLen := big.NewInt(int64(len(charsetStr)))
	password := make([]byte, length)
	for i := range password {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random number: %w", err)
		}
		password[i] = charsetStr[n.Int64()]
	}

	passwordStr := string(password)
	if len(passwordStr) == 0 {
		return nil, fmt.Errorf("generated password is empty")
	}

	return map[string]string{
		"value":       passwordStr,
		"charset":     charsetStr,
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"length":      fmt.Sprintf("%d", length),
	}, nil
}
