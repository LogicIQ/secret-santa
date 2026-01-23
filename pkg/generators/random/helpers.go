package random

func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

func getIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	switch val := config[key].(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	}
	return defaultValue
}

func getBoolConfig(config map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultValue
}

func getStringSliceConfig(config map[string]interface{}, key string) []string {
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
