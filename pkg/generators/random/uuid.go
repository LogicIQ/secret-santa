package random

import (
	"github.com/google/uuid"
	"time"
)

type UUIDGenerator struct{}

func (g *UUIDGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	id := uuid.New()
	return map[string]string{
		"value":       id.String(),
		"version":     "4",
		"variant":     "RFC4122",
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}
