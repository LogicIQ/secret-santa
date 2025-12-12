package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// ECDSAKeyGenerator generates ECDSA key pairs
type ECDSAKeyGenerator struct{}

func (g *ECDSAKeyGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	curve := getStringConfig(config, "curve", "P256")

	var ellipticCurve elliptic.Curve
	switch curve {
	case "P224":
		ellipticCurve = elliptic.P224()
	case "P256":
		ellipticCurve = elliptic.P256()
	case "P384":
		ellipticCurve = elliptic.P384()
	case "P521":
		ellipticCurve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported curve: %s", curve)
	}

	// Generate ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(ellipticCurve, rand.Reader)
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
		"curve":              curve,
		"algorithm":          "ECDSA",
	}, nil
}
