package tls

import (
	"encoding/pem"
	"testing"
)

func TestPrivateKeyGenerator_Generate(t *testing.T) {
	gen := &PrivateKeyGenerator{}

	tests := []struct {
		name      string
		config    map[string]interface{}
		algorithm string
	}{
		{
			name:      "default RSA",
			config:    map[string]interface{}{},
			algorithm: "RSA",
		},
		{
			name: "RSA 4096",
			config: map[string]interface{}{
				"algorithm": "RSA",
				"rsa_bits":  4096,
			},
			algorithm: "RSA",
		},
		{
			name: "ECDSA P256",
			config: map[string]interface{}{
				"algorithm":   "ECDSA",
				"ecdsa_curve": "P256",
			},
			algorithm: "ECDSA",
		},
		{
			name: "ED25519",
			config: map[string]interface{}{
				"algorithm": "ED25519",
			},
			algorithm: "ED25519",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			expectedKeys := []string{"private_key_pem", "public_key_pem", "public_key_openssh", "public_key_fingerprint_md5", "public_key_fingerprint_sha256"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			// Validate PEM format
			block, _ := pem.Decode([]byte(result["private_key_pem"]))
			if block == nil {
				t.Error("Generate() invalid private key PEM")
			}
		})
	}
}
