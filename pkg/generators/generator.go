package generators

import "strings"

// Generator interface for all secret generators
type Generator interface {
	Generate(config map[string]interface{}) (map[string]string, error)
}

func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if config == nil {
		return defaultValue
	}
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
	if config == nil {
		return strings.ToLower(defaultValue)
	}
	if val, ok := config[key].(string); ok {
		return strings.ToLower(strings.TrimSpace(val))
	}
	return strings.ToLower(defaultValue)
}

func getIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	if config == nil {
		return defaultValue
	}
	switch val := config[key].(type) {
	case int:
		return val
	case int64:
		return int(val)
	case int32:
		return int(val)
	case int16:
		return int(val)
	case int8:
		return int(val)
	case uint:
		if val > uint(^uint(0)>>1) {
			return defaultValue
		}
		return int(val)
	case uint64:
		if val > uint64(^uint(0)>>1) {
			return defaultValue
		}
		return int(val)
	case uint32:
		if uint(val) > uint(^uint(0)>>1) {
			return defaultValue
		}
		return int(val)
	case uint16:
		return int(val)
	case uint8:
		return int(val)
	case float64:
		if val >= float64(int(^uint(0)>>1)) || val <= float64(-int(^uint(0)>>1)-1) {
			return defaultValue
		}
		return int(val)
	case float32:
		if val >= float32(int(^uint(0)>>1)) || val <= float32(-int(^uint(0)>>1)-1) {
			return defaultValue
		}
		return int(val)
	}
	return defaultValue
}

// GetIntConfig exports the helper function for use in media packages
func GetIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	return getIntConfig(config, key, defaultValue)
}

func getBoolConfig(config map[string]interface{}, key string, defaultValue bool) bool {
	if config == nil {
		return defaultValue
	}
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultValue
}
