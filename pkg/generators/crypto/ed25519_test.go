package crypto

import (
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func TestED25519KeyGenerator_Generate(t *testing.T) {
	gen := &ED25519KeyGenerator{}

	result, err := gen.Generate(map[string]interface{}{})
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	expectedKeys := []string{"private_key_pem", "public_key_pem", "private_key_base64", "public_key_base64", "algorithm"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("Generate() missing key %s", key)
		}
	}

	if result["algorithm"] != "ED25519" {
		t.Errorf("Generate() algorithm = %s, want ED25519", result["algorithm"])
	}

	// Validate PEM format
	privateBlock, _ := pem.Decode([]byte(result["private_key_pem"]))
	if privateBlock == nil || privateBlock.Type != "PRIVATE KEY" {
		t.Error("Generate() invalid private key PEM format")
	}

	publicBlock, _ := pem.Decode([]byte(result["public_key_pem"]))
	if publicBlock == nil || publicBlock.Type != "PUBLIC KEY" {
		t.Error("Generate() invalid public key PEM format")
	}

	// Validate base64 encoding
	_, err = base64.StdEncoding.DecodeString(result["private_key_base64"])
	if err != nil {
		t.Errorf("Generate() invalid private key base64: %v", err)
	}

	_, err = base64.StdEncoding.DecodeString(result["public_key_base64"])
	if err != nil {
		t.Errorf("Generate() invalid public key base64: %v", err)
	}

	// Test multiple generations produce different keys
	result2, err := gen.Generate(map[string]interface{}{})
	if err != nil {
		t.Errorf("Generate() second call error = %v", err)
		return
	}

	if result["private_key_pem"] == result2["private_key_pem"] {
		t.Error("Generate() should produce different keys on each call")
	}
}