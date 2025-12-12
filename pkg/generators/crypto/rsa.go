package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// RSAKeyGenerator generates RSA key pairs for application use
type RSAKeyGenerator struct{}

func (g *RSAKeyGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	keySize := getIntConfig(config, "key_size", 2048)

	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
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
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
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
		"key_size":           fmt.Sprintf("%d", keySize),
		"algorithm":          "RSA",
	}, nil
}
