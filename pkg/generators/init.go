package generators

import (
	"github.com/logicIQ/secret-santa/pkg/generators/crypto"
	"github.com/logicIQ/secret-santa/pkg/generators/random"
	timegens "github.com/logicIQ/secret-santa/pkg/generators/time"
	"github.com/logicIQ/secret-santa/pkg/generators/tls"
)

func init() {
	Register("tls_private_key", &tls.PrivateKeyGenerator{})
	Register("tls_self_signed_cert", &tls.SelfSignedCertGenerator{})
	Register("tls_cert_request", &tls.CertRequestGenerator{})
	Register("tls_locally_signed_cert", &tls.LocallySignedCertGenerator{})

	Register("random_password", &random.PasswordGenerator{})
	Register("random_string", &random.StringGenerator{})
	Register("random_uuid", &random.UUIDGenerator{})
	Register("random_integer", &random.IntegerGenerator{})
	Register("random_bytes", &random.BytesGenerator{})
	Register("random_id", &random.IDGenerator{})

	Register("time_static", &timegens.StaticGenerator{})

	Register("crypto_hmac", &crypto.HMACGenerator{})
	Register("crypto_aes_key", &crypto.AESKeyGenerator{})
	Register("crypto_rsa_key", &crypto.RSAKeyGenerator{})
	Register("crypto_ed25519_key", &crypto.ED25519KeyGenerator{})
	Register("crypto_chacha20_key", &crypto.ChaCha20KeyGenerator{})
	Register("crypto_xchacha20_key", &crypto.XChaCha20KeyGenerator{})
	Register("crypto_ecdsa_key", &crypto.ECDSAKeyGenerator{})
	Register("crypto_ecdh_key", &crypto.ECDHKeyGenerator{})
}
