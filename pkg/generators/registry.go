package generators

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu         sync.RWMutex
	generators map[string]Generator
}

var globalRegistry = &Registry{
	generators: make(map[string]Generator),
}

func Register(generatorType string, generator Generator) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.generators[generatorType] = generator
}

func Get(generatorType string) (Generator, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	generator, exists := globalRegistry.generators[generatorType]
	if !exists {
		return nil, fmt.Errorf("unsupported generator type: %s", generatorType)
	}
	return generator, nil
}

func GetSupportedTypes() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	types := make([]string, 0, len(globalRegistry.generators))
	for generatorType := range globalRegistry.generators {
		types = append(types, generatorType)
	}
	return types
}

func IsSupported(generatorType string) bool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	_, exists := globalRegistry.generators[generatorType]
	return exists
}

func Clear() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.generators = make(map[string]Generator)
}
