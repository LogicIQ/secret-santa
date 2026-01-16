package generators

import "strings"

// Generator interface for all secret generators
type Generator interface {
	Generate(config map[string]interface{}) (map[string]string, error)
}

func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

// GetStringConfig exports the helper function for use in media packages
func GetStringConfig(config map[string]interface{}, key, defaultValue string) string {
	return getStringConfig(config, key, defaultValue)
}

// getNormalizedStringConfig gets a string config value and normalizes it to lowercase
func getNormalizedStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return strings.ToLower(strings.TrimSpace(val))
	}
	return strings.ToLower(defaultValue)
}

func getIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	switch val := config[key].(type) {
	case int:
		return val
	case float64:
		return int(val)
	}
	return defaultValue
}

// GetIntConfig exports the helper function for use in media packages
func GetIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	return getIntConfig(config, key, defaultValue)
}

func getBoolConfig(config map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultValue
}
