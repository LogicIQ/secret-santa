package template

import (
	"math"
	"strings"
	"testing"
)

func TestSHA256(t *testing.T) {
	result := SHA256("hello")
	if len(result) != 64 {
		t.Errorf("SHA256 should return 64 character hex string, got %d", len(result))
	}
}

func TestBcrypt(t *testing.T) {
	result, err := Bcrypt("password")
	if err != nil {
		t.Errorf("Bcrypt should not return error, got %v", err)
		return
	}
	if !strings.HasPrefix(result, "$2a$") {
		t.Errorf("Bcrypt should return hash starting with $2a$, got %s", result)
	}
}

func TestEntropy(t *testing.T) {
	result, err := Entropy("abc", "abcdefghijklmnopqrstuvwxyz")
	if err != nil {
		t.Errorf("Entropy should not return error, got %v", err)
		return
	}
	expected := 3 * math.Log2(26)
	if result < expected-0.1 || result > expected+0.1 {
		t.Errorf("Entropy should be approximately %.2f, got %.2f", expected, result)
	}
}

func TestCRC32(t *testing.T) {
	result := CRC32("hello")
	if len(result) != 8 {
		t.Errorf("CRC32 should return 8 character hex string, got %d", len(result))
	}
}

func TestURLSafeBase64(t *testing.T) {
	result := URLSafeBase64("hello")
	expected := "aGVsbG8="
	if result != expected {
		t.Errorf("URLSafeBase64('hello') = %s, want %s", result, expected)
	}
}

func TestRemoveHyphens(t *testing.T) {
	result := RemoveHyphens("550e8400-e29b-41d4-a716-446655440000")
	expected := "550e8400e29b41d4a716446655440000"
	if result != expected {
		t.Errorf("RemoveHyphens() = %s, want %s", result, expected)
	}
}

func TestToBinary(t *testing.T) {
	tests := []struct {
		input       interface{}
		expected    string
		expectError bool
	}{
		{5, "101", false},
		{"5", "101", false},
		{"invalid", "", true},
	}

	for _, test := range tests {
		result, err := ToBinary(test.input)
		if test.expectError {
			if err == nil {
				t.Errorf("ToBinary(%v) expected error, got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("ToBinary(%v) unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("ToBinary(%v) = %s, want %s", test.input, result, test.expected)
			}
		}
	}
}

func TestToHex(t *testing.T) {
	tests := []struct {
		input       interface{}
		expected    string
		expectError bool
	}{
		{255, "ff", false},
		{"255", "ff", false},
		{"invalid", "", true},
	}

	for _, test := range tests {
		result, err := ToHex(test.input)
		if test.expectError {
			if err == nil {
				t.Errorf("ToHex(%v) expected error, got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("ToHex(%v) unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("ToHex(%v) = %s, want %s", test.input, result, test.expected)
			}
		}
	}
}
