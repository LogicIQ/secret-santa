# Template Functions Optimization

## Overview

Moved from exporting computed values in generators to using Go template functions for better maintainability and flexibility.

## Changes Made

### 1. Created Template Functions Package (`pkg/template/`)

**New Functions:**
- `sha256` - SHA256 hash of input string
- `bcrypt` - Bcrypt hash of input string  
- `entropy` - Calculate entropy bits for string + charset
- `crc32` - CRC32 checksum of input string
- `urlSafeB64` - URL-safe base64 encoding
- `compact` - Remove hyphens from string
- `toBinary` - Convert integer/string to binary
- `toHex` - Convert integer/string to hexadecimal

### 2. Simplified Generator Outputs

**Before (redundant exports):**
```go
return map[string]string{
    "value":       string(password),
    "bcryptHash":  string(bcryptHash),
    "length":      fmt.Sprintf("%d", len(password)),
    "entropy":     fmt.Sprintf("%.2f", entropy),
    "sha256":      hex.EncodeToString(sha256Hash[:]),
    "base64":      base64.StdEncoding.EncodeToString(result),
    "urlSafe":     base64.URLEncoding.EncodeToString(result),
    "hex":         fmt.Sprintf("%x", result),
    "binary":      fmt.Sprintf("%b", result),
    "compact":     strings.ReplaceAll(id.String(), "-", ""),
}
```

**After (minimal exports):**
```go
return map[string]string{
    "value":       string(password),
    "charset":     charsetStr,
    "generatedAt": time.Now().UTC().Format(time.RFC3339),
}
```

### 3. Updated Template Usage

**Before:**
```yaml
template: |
  password: {{ .Password.value }}
  bcrypt_hash: {{ .Password.bcryptHash }}
  length: {{ .Password.length }}
  entropy: {{ .Password.entropy }}
  sha256: {{ .Password.sha256 }}
```

**After:**
```yaml
template: |
  password: {{ .Password.value }}
  bcrypt_hash: {{ .Password.value | bcrypt }}
  length: {{ len .Password.value }}
  entropy: {{ entropy .Password.value .Password.charset | printf "%.2f" }}
  sha256: {{ .Password.value | sha256 }}
```

## Benefits

### Memory Efficiency
- **50-80% reduction** in exported values per generator
- Only essential values exported (value, metadata, timestamps)
- Computed values generated on-demand in templates

### Maintainability
- Template functions centralized in `pkg/template/`
- No duplication of encoding/hashing logic across generators
- Single source of truth for transformations

### Flexibility
- Functions work with any string input
- Composable with Sprig functions: `{{ .Value | sha256 | upper }}`
- Conditional logic: `{{ if gt (entropy .Password.value .Password.charset) 60.0 }}`

### Consistency
- Same transformation approach across all generators
- Uniform function naming and behavior
- Better template readability

## Template Function Examples

```yaml
# Encoding transformations
api_key_b64: {{ .APIKey.value | b64enc }}
api_key_url_safe: {{ .APIKey.value | urlSafeB64 }}

# Numeric representations  
port_hex: {{ .Port.value | toHex }}
port_binary: {{ .Port.value | toBinary }}

# String transformations
uuid_compact: {{ .UUID.value | compact }}
password_hash: {{ .Password.value | sha256 }}

# Security calculations
entropy_bits: {{ entropy .Password.value .Password.charset | printf "%.1f" }}
checksum: {{ .Data.value | crc32 }}

# Conditional logic
{{- if gt (entropy .Password.value .Password.charset) 60.0 }}
strength: "high"
{{- else }}
strength: "medium"
{{- end }}

# Function composition
secure_hash: {{ .Password.value | sha256 | upper | trunc 16 }}
```

## Migration Guide

### For Existing Templates

Replace exported field access with function calls:

| Old | New |
|-----|-----|
| `{{ .Gen.length }}` | `{{ len .Gen.value }}` |
| `{{ .Gen.base64 }}` | `{{ .Gen.value \| b64enc }}` |
| `{{ .Gen.urlSafe }}` | `{{ .Gen.value \| urlSafeB64 }}` |
| `{{ .Gen.hex }}` | `{{ .Gen.value \| toHex }}` |
| `{{ .Gen.binary }}` | `{{ .Gen.value \| toBinary }}` |
| `{{ .Gen.sha256 }}` | `{{ .Gen.value \| sha256 }}` |
| `{{ .Gen.bcryptHash }}` | `{{ .Gen.value \| bcrypt }}` |
| `{{ .Gen.compact }}` | `{{ .Gen.value \| compact }}` |

### For New Templates

Use template functions from the start for maximum flexibility and performance.

## Testing

- All template functions have unit tests in `pkg/template/functions_test.go`
- Generator tests updated to match new minimal output format
- E2E tests updated to use template functions
- All tests passing ✅