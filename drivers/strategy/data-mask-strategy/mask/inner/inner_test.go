package inner

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockMaskFunc(input string) string {
	if input == "" {
		return ""
	}
	return "MASKED"
}

func TestNameMaskDriver(t *testing.T) {
	driver := NewNameMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{`{"name": "John Doe"}`, `{"name":"MASKED"}`},
		{`{"cname": "Jane Doe"}`, `{"cname":"MASKED"}`},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.JSONEq(t, tt.output, string(masked))
	}
}

func TestPhoneMaskDriver(t *testing.T) {
	driver := newPhoneMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{`12345678901`, "12345678901"},
		{`13456278901`, "MASKED"},
		{`19876543210`, "MASKED"},
		{`198765432102`, "198765432102"},
		{`19876543210a`, "19876543210a"},

		{"normal text", "normal text"},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.Equal(t, tt.output, string(masked))
	}
}

func TestIDCardMaskDriver(t *testing.T) {
	driver := newIDCardMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{`12345678901234567X`, "12345678901234567X"}, // Assuming valid ID for mock
		{`invalid-idformat`, "invalid-idformat"},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.Equal(t, tt.output, string(masked))
	}
}

func TestBankCardMaskDriver(t *testing.T) {
	driver := newBankCardMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{`1234567890123456`, "1234567890123456"}, // Assuming valid card number for mock
		{`invalid-card`, "invalid-card"},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.Equal(t, tt.output, string(masked))
	}
}

func TestAmountMaskDriver(t *testing.T) {
	driver := newAmountMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{`123.45`, "MASKED"}, // Assuming this is valid for testing
		{`not-amount`, "not-amount"},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.Equal(t, tt.output, string(masked))
	}
}

func TestDateMaskDriver(t *testing.T) {
	driver := newDateMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{`2023-01-01`, "MASKED"}, // Assuming this is valid for testing
		{`not-date`, "not-date"},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.Equal(t, tt.output, string(masked))
	}
}

func TestNameMaskDriver_LongText(t *testing.T) {
	driver := NewNameMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{`{"name": "Johnathon Maximillian Doe the Third"}`, `{"name":"MASKED"}`},
		{`{"normal_text": "This is a sentence with a name Johnathan Does embedded"}`, `{"normal_text":"This is a sentence with a name Johnathan Does embedded"}`},
		{`{"name": "", "cname": ""}`, `{"name":"", "cname":""}`},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.JSONEq(t, tt.output, string(masked))
	}
}

func TestPhoneMaskDriver_LongText(t *testing.T) {
	driver := newPhoneMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{strings.Repeat("1", 50) + "3987654321" + strings.Repeat("0", 50), strings.Repeat("1", 50) + "3987654321" + strings.Repeat("0", 50)},
		{"Contact numbers: 13987654321, 15800001111", "Contact numbers: MASKED, MASKED"},
		{strings.Repeat("normal-text", 100), strings.Repeat("normal-text", 100)},
		{"Dialing sequence: 08005555555", "Dialing sequence: 08005555555"},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.Equal(t, tt.output, string(masked))
	}
}

func TestIDCardMaskDriver_LongText(t *testing.T) {
	driver := newIDCardMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{strings.Repeat("1", 50) + "12345678901234567X" + strings.Repeat("X", 50), strings.Repeat("1", 50) + "12345678901234567X" + strings.Repeat("X", 50)},
		{strings.Repeat("normal-text", 100), strings.Repeat("normal-text", 100)},
		{"Incorrect format: 123-456-789-00", "Incorrect format: 123-456-789-00"},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.Equal(t, tt.output, string(masked))
	}
}

func TestBankCardMaskDriver_LongText(t *testing.T) {
	driver := newBankCardMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{strings.Repeat("1", 50) + "1234567890123456" + strings.Repeat("9", 50), strings.Repeat("1", 50) + "1234567890123456" + strings.Repeat("9", 50)},
		{"Valid card numbers: dasdw, 8765432187654321", "Valid card numbers: MASKED, MASKED"},
		{strings.Repeat("normal-text", 100), strings.Repeat("normal-text", 100)},
		{"Fake card number: 1111-2222-3333-4444", "Fake card number: 1111-2222-3333-4444"},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.Equal(t, tt.output, string(masked))
	}
}

func TestAmountMaskDriver_LongText(t *testing.T) {
	driver := newAmountMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{"Invoice total: $1234567890.99", "MASKED"},
		{"Transaction amounts: $12345.67, fee: $0.02", "MASKED, fee: MASKED"},
		{strings.Repeat("normal-text", 100), strings.Repeat("normal-text", 100)},
		{"Plain text values: twenty five dollars", "Plain text values: twenty five dollars"},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.Equal(t, tt.output, string(masked))
	}
}

func TestDateMaskDriver_LongText(t *testing.T) {
	driver := newDateMaskDriver(mockMaskFunc)

	tests := []struct {
		input  string
		output string
	}{
		{"Event date is 2077-12-31 extended overlap text", "MASKED extended overlap text"},
		{"Date strings: 2025-09-24T13:45:00, obsolete: 1912-04-15T00:00:00", "MASKED, obsolete: MASKED"},
		{strings.Repeat("normal-text", 100), strings.Repeat("normal-text", 100)},
		{"Formatted text: the year nineteen ninety nine", "Formatted text: the year nineteen ninety nine"},
	}

	for _, tt := range tests {
		masked, err := driver.Exec([]byte(tt.input))
		assert.NoError(t, err)
		assert.Equal(t, tt.output, string(masked))
	}
}
