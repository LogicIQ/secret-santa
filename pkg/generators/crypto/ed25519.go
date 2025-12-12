package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
)

// ED25519KeyGenerator generates ED25519 key pairs for application use
type ED25519KeyGenerator struct{}

func (g *ED25519KeyGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	// Generate ED25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Encode private key (PKCS#8)
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return map[string]string{
		"private_key_pem":    string(privateKeyPEM),
		"public_key_pem":     string(publicKeyPEM),
		"private_key_base64": base64.StdEncoding.EncodeToString(privateKeyPEM),
		"public_key_base64":  base64.StdEncoding.EncodeToString(publicKeyPEM),
		"algorithm":          "ED25519",
	}, nil
}
