package crypto

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
	case float64:
		return int(val)
	}
	return defaultValue
}
