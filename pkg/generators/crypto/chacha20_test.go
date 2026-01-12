package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestChaCha20KeyGenerator_Generate(t *testing.T) {
	gen := &ChaCha20KeyGenerator{}

	result, err := gen.Generate(map[string]interface{}{})
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	expectedKeys := []string{"key_base64", "key_hex", "key_size", "algorithm"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("Generate() missing key %s", key)
		}
	}

	// Validate key size
	if result["key_size"] != "256" {
		t.Errorf("Generate() key_size = %s, want 256", result["key_size"])
		return
	}

	if result["algorithm"] != "ChaCha20" {
		t.Errorf("Generate() algorithm = %s, want ChaCha20", result["algorithm"])
		return
	}

	// Validate base64 key length (32 bytes = 44 base64 chars with padding)
	decoded, err := base64.StdEncoding.DecodeString(result["key_base64"])
	if err != nil {
		t.Errorf("Generate() invalid base64: %v", err)
		return
	}
	if len(decoded) != 32 {
		t.Errorf("Generate() key length = %d, want 32", len(decoded))
	}

	// Validate hex key length (32 bytes = 64 hex chars)
	hexDecoded, err := hex.DecodeString(result["key_hex"])
	if err != nil {
		t.Errorf("Generate() invalid hex: %v", err)
		return
	}
	if len(hexDecoded) != 32 {
		t.Errorf("Generate() hex key length = %d, want 32", len(hexDecoded))
	}
}

func TestXChaCha20KeyGenerator_Generate(t *testing.T) {
	gen := &XChaCha20KeyGenerator{}

	result, err := gen.Generate(map[string]interface{}{})
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	expectedKeys := []string{"key_base64", "key_hex", "key_size", "algorithm"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("Generate() missing key %s", key)
		}
	}

	// Validate key size
	if result["key_size"] != "256" {
		t.Errorf("Generate() key_size = %s, want 256", result["key_size"])
		return
	}

	if result["algorithm"] != "XChaCha20" {
		t.Errorf("Generate() algorithm = %s, want XChaCha20", result["algorithm"])
		return
	}

	// Validate base64 key length (32 bytes)
	decoded, err := base64.StdEncoding.DecodeString(result["key_base64"])
	if err != nil {
		t.Errorf("Generate() invalid base64: %v", err)
		return
	}
	if len(decoded) != 32 {
		t.Errorf("Generate() key length = %d, want 32", len(decoded))
	}

	// Validate hex key length (32 bytes)
	hexDecoded, err := hex.DecodeString(result["key_hex"])
	if err != nil {
		t.Errorf("Generate() invalid hex: %v", err)
		return
	}
	if len(hexDecoded) != 32 {
		t.Errorf("Generate() hex key length = %d, want 32", len(hexDecoded))
	}
}
