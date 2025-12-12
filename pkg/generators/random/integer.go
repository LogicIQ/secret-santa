package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// IntegerGenerator generates random integers
type IntegerGenerator struct{}

func (g *IntegerGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	min := getIntConfig(config, "min", 0)
	max := getIntConfig(config, "max", 100)

	if min > max {
		return nil, fmt.Errorf("min (%d) cannot be greater than max (%d)", min, max)
	}

	// Generate random number in range [min, max]
	rangeSize := max - min + 1
	n, err := rand.Int(rand.Reader, big.NewInt(int64(rangeSize)))
	if err != nil {
		return nil, err
	}

	result := min + int(n.Int64())

	return map[string]string{
		"value":       fmt.Sprintf("%d", result),
		"min":         fmt.Sprintf("%d", min),
		"max":         fmt.Sprintf("%d", max),
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}
