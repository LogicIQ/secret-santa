package crypto

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

type ECDHKeyGenerator struct{}

func (g *ECDHKeyGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	curve := getStringConfig(config, "curve", "P256")

	var ecdhCurve ecdh.Curve
	switch curve {
	case "P256":
		ecdhCurve = ecdh.P256()
	case "P384":
		ecdhCurve = ecdh.P384()
	case "P521":
		ecdhCurve = ecdh.P521()
	case "X25519":
		ecdhCurve = ecdh.X25519()
	default:
		return nil, fmt.Errorf("unsupported curve: %s", curve)
	}

	// Generate ECDH key pair
	privateKey, err := ecdhCurve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.PublicKey()

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
		"private_key_base64": base64.StdEncoding.EncodeToString(privateKey.Bytes()),
		"public_key_base64":  base64.StdEncoding.EncodeToString(publicKey.Bytes()),
		"curve":              curve,
		"algorithm":          "ECDH",
	}, nil
}
