package keyword

import (
	"strings"
	"testing"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
)

func TestKeywordDriver(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *mask.Rule
		maskFunc mask.MaskFunc
		input    string
		expected string
	}{
		{
			name: "Basic replacement",
			cfg: &mask.Rule{
				Match: &mask.BasicItem{
					Type:  mask.MatchKeyword,
					Value: "secret",
				},
				Mask: &mask.Mask{},
			},
			maskFunc: func(value string) string {
				return "****"
			},
			input:    "This is a secret text.",
			expected: "This is a **** text.",
		},
		{
			name: "No replacement needed",
			cfg: &mask.Rule{
				Match: &mask.BasicItem{
					Type:  mask.MatchKeyword,
					Value: "hidden",
				},
				Mask: &mask.Mask{},
			},
			maskFunc: func(value string) string {
				return "****"
			},
			input:    "This is a visible text.",
			expected: "This is a visible text.",
		},
		{
			name: "Multiple occurrences",
			cfg: &mask.Rule{
				Match: &mask.BasicItem{
					Type:  mask.MatchKeyword,
					Value: "cat",
				},
				Mask: &mask.Mask{},
			},
			maskFunc: func(value string) string {
				return "dog"
			},
			input:    "A cat chasing another cat.",
			expected: "A dog chasing another dog.",
		},
		{
			name: "Empty input",
			cfg: &mask.Rule{
				Match: &mask.BasicItem{
					Type:  mask.MatchKeyword,
					Value: "anything",
				},
				Mask: &mask.Mask{},
			},
			maskFunc: func(value string) string {
				return "nothing"
			},
			input:    "",
			expected: "",
		},
		{
			name: "Long text replacement",
			cfg: &mask.Rule{
				Match: &mask.BasicItem{
					Type:  mask.MatchKeyword,
					Value: "important",
				},
				Mask: &mask.Mask{},
			},
			maskFunc: func(value string) string {
				return "[REDACTED]"
			},
			input:    strings.Repeat("This is important. ", 1000),
			expected: strings.Repeat("This is [REDACTED]. ", 1000),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			driver, err := NewKeywordMaskDriver(test.cfg, test.maskFunc)
			if err != nil {
				t.Fatalf("Failed to create driver: %v", err)
			}

			result, _ := driver.Exec([]byte(test.input))
			if string(result) != test.expected {
				t.Errorf("Expected '%v', but got '%v'", test.expected, string(result))
			}
		})
	}
}
