package tls

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
)

const (
	KeyAlgorithmRSA     = "RSA"
	KeyAlgorithmECDSA   = "ECDSA"
	KeyAlgorithmED25519 = "ED25519"
	KeyAlgorithmUnknown = "UNKNOWN"
)

func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

func getIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key].(int); ok {
		return val
	}
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}

// getStringSliceConfig extracts a string slice from config map.
// Non-string elements in the slice are silently ignored, which may result in partial results.
// Returns nil if the key doesn't exist or the value is not a slice.
func getStringSliceConfig(config map[string]interface{}, key string) []string {
	if config == nil {
		return nil
	}
	if val, ok := config[key].([]interface{}); ok {
		var result []string
		for _, v := range val {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return nil
}

func getKeyAlgorithm(privateKey interface{}) string {
	switch privateKey.(type) {
	case *rsa.PrivateKey:
		return KeyAlgorithmRSA
	case *ecdsa.PrivateKey:
		return KeyAlgorithmECDSA
	case ed25519.PrivateKey:
		return KeyAlgorithmED25519
	default:
		return KeyAlgorithmUnknown
	}
}

func publicKeysMatch(privateKey, publicKey interface{}) bool {
	switch priv := privateKey.(type) {
	case *rsa.PrivateKey:
		pub, ok := publicKey.(*rsa.PublicKey)
		return ok && priv.PublicKey.N.Cmp(pub.N) == 0 && priv.PublicKey.E == pub.E
	case *ecdsa.PrivateKey:
		pub, ok := publicKey.(*ecdsa.PublicKey)
		return ok && priv.PublicKey.Curve == pub.Curve && priv.PublicKey.X.Cmp(pub.X) == 0 && priv.PublicKey.Y.Cmp(pub.Y) == 0
	case ed25519.PrivateKey:
		pub, ok := publicKey.(ed25519.PublicKey)
		return ok && priv.Public().(ed25519.PublicKey).Equal(pub)
	default:
		return false
	}
}
