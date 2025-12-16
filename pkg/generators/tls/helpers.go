package tls

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
)


func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

func getIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}

func getStringSliceConfig(config map[string]interface{}, key string) []string {
	if val, ok := config[key].([]interface{}); ok {
		result := make([]string, len(val))
		for i, v := range val {
			if str, ok := v.(string); ok {
				result[i] = str
			}
		}
		return result
	}
	return nil
}

func getKeyAlgorithm(privateKey interface{}) string {
	switch privateKey.(type) {
	case *rsa.PrivateKey:
		return "RSA"
	case *ecdsa.PrivateKey:
		return "ECDSA"
	case ed25519.PrivateKey:
		return "ED25519"
	default:
		return "UNKNOWN"
	}
}
