package mask

import (
	"strings"
	"testing"
)

func TestPartialDisplay(t *testing.T) {
	tests := []struct {
		origin   string
		begin    int
		length   int
		expected string
	}{
		{"abcdef", 2, 3, "**cde*"},
		{"abcdef", 0, 4, "abcd**"},
		{"abcdef", 10, 2, "******"},
		{"abcdef", 2, -1, "**cdef"},
	}
	for _, test := range tests {
		f := partialDisplay(test.begin, test.length)
		result := f(test.origin)
		if result != test.expected {
			t.Errorf("Expected %s but got %s", test.expected, result)
		}
	}
}

func TestPartialMask(t *testing.T) {
	tests := []struct {
		origin   string
		begin    int
		length   int
		expected string
	}{
		{"abcdef", 2, 3, "ab***f"},
		{"abcdef", 0, 4, "****ef"},
		{"abcdef", 10, 2, "abcdef"},
		{"abcdef", 2, -1, "ab****"},
	}
	for _, test := range tests {
		f := partialMasking(test.begin, test.length)
		result := f(test.origin)
		if result != test.expected {
			t.Errorf("Expected %s but got %s", test.expected, result)
		}
	}
}

func TestTruncation(t *testing.T) {
	tests := []struct {
		origin   string
		begin    int
		length   int
		expected string
	}{
		{"abcdef", 2, 3, "cde"},
		{"abcdef", 0, 4, "abcd"},
		{"abcdef", 10, 2, ""},
		{"abcdef", 2, -1, "cdef"},
	}
	for _, test := range tests {
		f := truncation(test.begin, test.length)
		result := f(test.origin)
		if result != test.expected {
			t.Errorf("Expected %s but got %s", test.expected, result)
		}
	}
}

func TestReplacement(t *testing.T) {
	randomFunc, _ := replacement(ReplaceRandom, "")
	customFunc, _ := replacement(ReplaceCustom, "custom")

	tests := []struct {
		origin   string
		maskFunc MaskFunc
		funcType string
		expected string
	}{
		{"abcdef", randomFunc, ReplaceRandom, ""},
		{"abcdef", customFunc, ReplaceCustom, "custom"},
	}
	for _, test := range tests {
		result := test.maskFunc(test.origin)
		// For random, just check length
		if test.funcType == ReplaceRandom {
			if len(result) != len(test.origin) {
				t.Errorf("Expected length %d but got %d", len(test.origin), len(result))
			}
		} else {
			if result != test.expected {
				t.Errorf("Expected %s but got %s", test.expected, result)
			}
		}
	}
}

func TestShuffling(t *testing.T) {
	origin := "abcdef"
	f := shuffling(0, len(origin))
	result := f(origin)
	if len(result) != len(origin) {
		t.Errorf("Expected length %d but got %d", len(origin), len(result))
	}

	// Check that all original characters are present
	for _, char := range origin {
		if strings.Count(result, string(char)) != strings.Count(origin, string(char)) {
			t.Errorf("Character %c count mismatch", char)
		}
	}
}
