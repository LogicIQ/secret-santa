package tls

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"golang.org/x/crypto/ssh"
)

type PrivateKeyGenerator struct{}

func (g *PrivateKeyGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	algorithm := strings.ToUpper(strings.TrimSpace(getStringConfig(config, "algorithm", "RSA")))

	switch algorithm {
	case "RSA":
		return g.generateRSA(config)
	case "ECDSA":
		return g.generateECDSA(config)
	case "ED25519":
		return g.generateED25519(config)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s (supported: RSA, ECDSA, ED25519)", algorithm)
	}
}

func (g *PrivateKeyGenerator) generateRSA(config map[string]interface{}) (map[string]string, error) {
	bits := getIntConfig(config, "rsa_bits", 2048)

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	// PEM format
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// PKCS8 format
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyPKCS8 := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	// Public key PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// OpenSSH format
	sshPublicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"private_key_pem":               string(privateKeyPEM),
		"private_key_pem_pkcs8":         string(privateKeyPKCS8),
		"public_key_pem":                string(publicKeyPEM),
		"public_key_openssh":            string(ssh.MarshalAuthorizedKey(sshPublicKey)),
		"public_key_fingerprint_md5":    ssh.FingerprintLegacyMD5(sshPublicKey),
		"public_key_fingerprint_sha256": ssh.FingerprintSHA256(sshPublicKey),
	}, nil
}

func (g *PrivateKeyGenerator) generateECDSA(config map[string]interface{}) (map[string]string, error) {
	curve := strings.ToUpper(strings.TrimSpace(getStringConfig(config, "ecdsa_curve", "P224")))

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
		return nil, fmt.Errorf("unsupported ECDSA curve: %s", curve)
	}

	privateKey, err := ecdsa.GenerateKey(ellipticCurve, rand.Reader)
	if err != nil {
		return nil, err
	}

	// PKCS8 format
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	// Public key PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return map[string]string{
		"private_key_pem":       string(privateKeyPEM),
		"private_key_pem_pkcs8": string(privateKeyPEM),
		"public_key_pem":        string(publicKeyPEM),
	}, nil
}

func (g *PrivateKeyGenerator) generateED25519(config map[string]interface{}) (map[string]string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// PKCS8 format
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	// Public key PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	publicKeyPEMBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return map[string]string{
		"private_key_pem":       string(privateKeyPEM),
		"private_key_pem_pkcs8": string(privateKeyPEM),
		"public_key_pem":        string(publicKeyPEMBytes),
	}, nil
}
