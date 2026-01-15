package mapstructure

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_convertInt(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int
		wantError bool
	}{
		// Native int
		{"int positive", 42, 42, false},
		{"int negative", -42, -42, false},
		{"int zero", 0, 0, false},
		{"int64", int64(100), 100, false},
		// Native uint
		{"uint", uint(42), 42, false},
		{"uint64", uint64(100), 100, false},
		// Native float
		{"float positive", 42.9, 42, false},
		{"float negative", -42.9, -42, false},
		// Native bool
		{"bool true", true, 1, false},
		{"bool false", false, 0, false},
		// String parsing
		{"string valid", "42", 42, false},
		{"string negative", "-42", -42, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, int(result.Int()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertInt8(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int8
		wantError bool
	}{
		// Native types
		{"int positive", 127, 127, false},
		{"int negative", -128, -128, false},
		{"uint", uint(100), 100, false},
		{"float", 50.9, 50, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "127", 127, false},
		{"string negative", "-128", -128, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "128", 0, true},
		{"string underflow", "-129", 0, true},
		{"string invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt8(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, int8(result.Int()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertInt16(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int16
		wantError bool
	}{
		// Native types
		{"int positive", 32767, 32767, false},
		{"int negative", -32768, -32768, false},
		{"uint", uint(30000), 30000, false},
		{"float", 1000.9, 1000, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "32767", 32767, false},
		{"string negative", "-32768", -32768, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "32768", 0, true},
		{"string underflow", "-32769", 0, true},
		{"string invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt16(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, int16(result.Int()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertInt32(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int32
		wantError bool
	}{
		// Native types
		{"int positive", 1000000, 1000000, false},
		{"int negative", -1000000, -1000000, false},
		{"uint", uint(1000000), 1000000, false},
		{"float", 1000000.9, 1000000, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "2147483647", 2147483647, false},
		{"string negative", "-2147483648", -2147483648, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "2147483648", 0, true},
		{"string underflow", "-2147483649", 0, true},
		{"string invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt32(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, int32(result.Int()))
		})
	}
}

func TestConverter_convertInt64(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int64
		wantError bool
	}{
		// Native types
		{"int positive", 1000000, 1000000, false},
		{"int negative", -1000000, -1000000, false},
		{"int64", int64(9223372036854775807), 9223372036854775807, false},
		{"uint", uint(1000000), 1000000, false},
		{"float", 1000000.9, 1000000, false},
		{"bool true", true, 1, false},
		{"bool false", false, 0, false},
		// String
		{"string valid", "9223372036854775807", 9223372036854775807, false},
		{"string negative", "-9223372036854775808", -9223372036854775808, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "9223372036854775808", 0, true},
		{"string underflow", "-9223372036854775809", 0, true},
		{"string invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt64(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Int())
		})
	}
}

func TestConverter_convertUint(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint
		wantError bool
	}{
		// Native int
		{"int positive", 42, 42, false},
		{"int zero", 0, 0, false},
		{"int negative", -1, 0, true},
		// Native uint
		{"uint", uint(42), 42, false},
		{"uint64", uint64(100), 100, false},
		// Native float
		{"float positive", 42.9, 42, false},
		{"float negative", -1.5, 0, true},
		// Native bool
		{"bool true", true, 1, false},
		{"bool false", false, 0, false},
		// String parsing
		{"string valid", "42", 42, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string negative", "-1", 0, true},
		{"string invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, uint(result.Uint()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertUint8(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint8
		wantError bool
	}{
		// Native types
		{"int positive", 255, 255, false},
		{"int negative", -1, 0, true},
		{"uint", uint(200), 200, false},
		{"float", 100.9, 100, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "255", 255, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "256", 0, true},
		{"string negative", "-1", 0, true},
		{"string invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint8(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, uint8(result.Uint()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertUint16(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint16
		wantError bool
	}{
		// Native types
		{"int positive", 65535, 65535, false},
		{"int negative", -1, 0, true},
		{"uint", uint(50000), 50000, false},
		{"float", 1000.9, 1000, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "65535", 65535, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "65536", 0, true},
		{"string negative", "-1", 0, true},
		{"string invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint16(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, uint16(result.Uint()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertUint32(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint32
		wantError bool
	}{
		// Native types
		{"int positive", 1000000, 1000000, false},
		{"int negative", -1, 0, true},
		{"uint", uint(4294967295), 4294967295, false},
		{"float", 1000000.9, 1000000, false},
		{"bool true", true, 1, false},
		// String
		{"string valid", "4294967295", 4294967295, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "4294967296", 0, true},
		{"string negative", "-1", 0, true},
		{"string invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint32(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, uint32(result.Uint()))
		})
	}
}

func TestConverter_convertUint64(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint64
		wantError bool
	}{
		// Native types
		{"int positive", 1000000, 1000000, false},
		{"int negative", -1, 0, true},
		{"uint", uint(1000000), 1000000, false},
		{"uint64", uint64(18446744073709551615), 18446744073709551615, false},
		{"float", 1000000.9, 1000000, false},
		{"bool true", true, 1, false},
		{"bool false", false, 0, false},
		// String
		{"string valid", "18446744073709551615", 18446744073709551615, false},
		{"string zero", "0", 0, false},
		{"string empty", "", 0, false},
		{"string overflow", "18446744073709551616", 0, true},
		{"string negative", "-1", 0, true},
		{"string invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint64(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Uint())
		})
	}
}

func TestConverter_convertFloat32(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      float32
		wantError bool
	}{
		// Native int
		{"int positive", 42, 42.0, false},
		{"int negative", -42, -42.0, false},
		{"int zero", 0, 0.0, false},
		{"int64", int64(100), 100.0, false},
		// Native uint
		{"uint", uint(42), 42.0, false},
		{"uint64", uint64(100), 100.0, false},
		// Native float
		{"float32", float32(3.14), 3.14, false},
		{"float64", 3.14, 3.14, false},
		// Native bool
		{"bool true", true, 1.0, false},
		{"bool false", false, 0.0, false},
		// String parsing
		{"string positive", "3.14", 3.14, false},
		{"string negative", "-3.14", -3.14, false},
		{"string zero", "0.0", 0.0, false},
		{"string integer", "42", 42.0, false},
		{"string scientific", "1e10", 1e10, false},
		{"string empty", "", 0.0, false},
		{"string invalid", "invalid", 0, true},
		{"string infinity", "+Inf", float32(math.Inf(1)), false},
		{"string negative infinity", "-Inf", float32(math.Inf(-1)), false},
		{"string NaN", "NaN", float32(math.NaN()), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertFloat32(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			got := float32(result.Float())
			switch tt.name {
			case "string NaN":
				assert.True(t, math.IsNaN(float64(got)))
			case "string infinity", "string negative infinity":
				assert.Equal(t, math.IsInf(float64(tt.want), 0), math.IsInf(float64(got), 0))
				if math.IsInf(float64(tt.want), 1) {
					assert.True(t, math.IsInf(float64(got), 1))
				} else if math.IsInf(float64(tt.want), -1) {
					assert.True(t, math.IsInf(float64(got), -1))
				}
			default:
				assert.InDelta(t, tt.want, got, 0.0001)
			}
		})
	}
}

func TestConverter_convertFloat64(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      float64
		wantError bool
	}{
		// Native int
		{"int positive", 42, 42.0, false},
		{"int negative", -42, -42.0, false},
		{"int zero", 0, 0.0, false},
		{"int64", int64(100), 100.0, false},
		// Native uint
		{"uint", uint(42), 42.0, false},
		{"uint64", uint64(100), 100.0, false},
		// Native float
		{"float32", float32(3.14), 3.14, false},
		{"float64", 3.141592653589793, 3.141592653589793, false},
		// Native bool
		{"bool true", true, 1.0, false},
		{"bool false", false, 0.0, false},
		// String parsing
		{"string positive", "3.141592653589793", 3.141592653589793, false},
		{"string negative", "-3.141592653589793", -3.141592653589793, false},
		{"string zero", "0.0", 0.0, false},
		{"string integer", "42", 42.0, false},
		{"string scientific", "1e100", 1e100, false},
		{"string empty", "", 0.0, false},
		{"string invalid", "invalid", 0, true},
		{"string infinity", "+Inf", math.Inf(1), false},
		{"string negative infinity", "-Inf", math.Inf(-1), false},
		{"string NaN", "NaN", math.NaN(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertFloat64(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			got := result.Float()
			switch tt.name {
			case "string NaN":
				assert.True(t, math.IsNaN(got))
			case "string infinity", "string negative infinity":
				assert.Equal(t, math.IsInf(tt.want, 0), math.IsInf(got, 0))
				if math.IsInf(tt.want, 1) {
					assert.True(t, math.IsInf(got, 1))
				} else if math.IsInf(tt.want, -1) {
					assert.True(t, math.IsInf(got, -1))
				}
			default:
				assert.InDelta(t, tt.want, got, 0.0001)
			}
		})
	}
}
