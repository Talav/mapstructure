package mapstructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_convertBool(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      bool
		wantError bool
	}{
		// Native bool
		{"bool true", true, true, false},
		{"bool false", false, false, false},
		// Native int
		{"int positive", 42, true, false},
		{"int zero", 0, false, false},
		{"int negative", -1, true, false},
		{"int8", int8(1), true, false},
		{"int64", int64(0), false, false},
		// Native uint
		{"uint positive", uint(42), true, false},
		{"uint zero", uint(0), false, false},
		{"uint8", uint8(1), true, false},
		// Native float
		{"float64 positive", 3.14, true, false},
		{"float64 zero", 0.0, false, false},
		{"float32", float32(1.5), true, false},
		// String parsing
		{"string true", "true", true, false},
		{"string false", "false", false, false},
		{"string TRUE", "TRUE", true, false},
		{"string 1", "1", true, false},
		{"string 0", "0", false, false},
		{"string empty", "", false, false},
		{"string invalid", "invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertBool(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Bool())
		})
	}
}

func TestConverter_convertString(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      string
		wantError bool
	}{
		// Native string
		{"string valid", "hello", "hello", false},
		{"string empty", "", "", false},
		{"string unicode", "世界", "世界", false},
		{"string with spaces", "hello world", "hello world", false},
		// Native bool
		{"bool true", true, "1", false},
		{"bool false", false, "0", false},
		// Native int
		{"int positive", 42, "42", false},
		{"int negative", -42, "-42", false},
		{"int zero", 0, "0", false},
		{"int64", int64(1234567890), "1234567890", false},
		// Native uint
		{"uint", uint(42), "42", false},
		{"uint64", uint64(1234567890), "1234567890", false},
		// Native float
		{"float positive", 3.14, "3.14", false},
		{"float negative", -3.14, "-3.14", false},
		{"float integer", 42.0, "42", false},
		// []byte
		{"bytes", []byte("hello"), "hello", false},
		{"bytes empty", []byte{}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertString(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.String())
		})
	}
}
