package random

import "math"

func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if config == nil {
		return defaultValue
	}
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

func getIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	if config == nil {
		return defaultValue
	}
	switch val := config[key].(type) {
	case int:
		return val
	case int64:
		if val > int64(math.MaxInt) || val < int64(math.MinInt) {
			return defaultValue
		}
		return int(val)
	case float64:
		if val > float64(math.MaxInt) || val < float64(math.MinInt) {
			return defaultValue
		}
		return int(val)
	}
	return defaultValue
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

func getStringSliceConfig(config map[string]interface{}, key string) []string {
	if config == nil {
		return nil
	}
	if val, ok := config[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return nil
}
