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

func FuncMap() template.FuncMap {
	funcMap := sprig.TxtFuncMap()
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

func SHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func Bcrypt(s string) string {
	password := []byte(s)
	if len(password) > 72 {
		password = password[:72]
	}
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

func Entropy(s, charset string) float64 {
	return float64(len(s)) * math.Log2(float64(len(charset)))
}

func CRC32(s string) string {
	return fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(s)))
}

func URLSafeBase64(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

func Compact(s string) string {
	return strings.ReplaceAll(s, "-", "")
}

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
