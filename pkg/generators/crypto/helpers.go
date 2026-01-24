package crypto

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
	case float64:
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return defaultValue
		}
		if val < math.MinInt || val > math.MaxInt {
			return defaultValue
		}
		if val != math.Trunc(val) {
			return defaultValue
		}
		return int(val)
	}
	return defaultValue
}
