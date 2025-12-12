package template

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"math"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"golang.org/x/crypto/bcrypt"
)

// FuncMap returns a template.FuncMap with all custom functions
func FuncMap() template.FuncMap {
	funcMap := sprig.TxtFuncMap()

	// Add custom functions
	funcMap["sha256"] = SHA256
	funcMap["bcrypt"] = Bcrypt
	funcMap["entropy"] = Entropy
	funcMap["crc32"] = CRC32
	funcMap["urlSafeB64"] = URLSafeBase64
	funcMap["compact"] = Compact
	funcMap["toBinary"] = ToBinary
	funcMap["toHex"] = ToHex

	return funcMap
}

// SHA256 returns the SHA256 hash of the input string
func SHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// Bcrypt returns the bcrypt hash of the input string
func Bcrypt(s string) string {
	password := []byte(s)
	if len(password) > 72 {
		password = password[:72]
	}
	hash, _ := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	return string(hash)
}

// Entropy calculates the entropy bits for a string given its charset
func Entropy(s, charset string) float64 {
	return float64(len(s)) * math.Log2(float64(len(charset)))
}

// CRC32 returns the CRC32 checksum of the input string
func CRC32(s string) string {
	return fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(s)))
}

// URLSafeBase64 returns the URL-safe base64 encoding of the input string
func URLSafeBase64(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

// Compact removes hyphens from the input string
func Compact(s string) string {
	return strings.ReplaceAll(s, "-", "")
}

// ToBinary converts an integer or string to binary representation
func ToBinary(i interface{}) string {
	switch v := i.(type) {
	case int:
		return fmt.Sprintf("%b", v)
	case string:
		if num, err := strconv.Atoi(v); err == nil {
			return fmt.Sprintf("%b", num)
		}
	}
	return ""
}

// ToHex converts an integer or string to hexadecimal representation
func ToHex(i interface{}) string {
	switch v := i.(type) {
	case int:
		return fmt.Sprintf("%x", v)
	case string:
		if num, err := strconv.Atoi(v); err == nil {
			return fmt.Sprintf("%x", num)
		}
	}
	return ""
}
