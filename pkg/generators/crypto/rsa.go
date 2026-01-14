package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

type RSAKeyGenerator struct{}

func (g *RSAKeyGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	keySize := getIntConfig(config, "key_size", 2048)

	// Validate key size
	if keySize < 2048 {
		return nil, fmt.Errorf("RSA key size too small, minimum 2048 bits, got: %d", keySize)
	}
	if keySize > 8192 {
		return nil, fmt.Errorf("RSA key size too large, maximum 8192 bits, got: %d", keySize)
	}
	if keySize%8 != 0 {
		return nil, fmt.Errorf("RSA key size must be a multiple of 8, got: %d", keySize)
	}

	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Encode private key (PKCS#8)
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})
	if privateKeyPEM == nil {
		return nil, fmt.Errorf("failed to encode private key PEM")
	}

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	if publicKeyPEM == nil {
		return nil, fmt.Errorf("failed to encode public key PEM")
	}

	return map[string]string{
		"private_key_pem":    string(privateKeyPEM),
		"public_key_pem":     string(publicKeyPEM),
		"private_key_base64": base64.StdEncoding.EncodeToString(privateKeyPEM),
		"public_key_base64":  base64.StdEncoding.EncodeToString(publicKeyPEM),
		"key_size":           fmt.Sprintf("%d", keySize),
		"algorithm":          "RSA",
	}, nil
}
