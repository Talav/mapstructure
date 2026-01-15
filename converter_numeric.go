package mapstructure

import (
	"fmt"
	"reflect"
	"strconv"
)

// convertInt converts a value to int.
func convertInt(value any) (reflect.Value, error) {
	i, err := convertToInt(value, 0)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(int(i)), nil
}

// convertInt8 converts a value to int8.
func convertInt8(value any) (reflect.Value, error) {
	i, err := convertToInt(value, 8)
	if err != nil {
		return reflect.Value{}, err
	}

	//nolint:gosec // Overflow checked by bitSize in convertToInt
	return reflect.ValueOf(int8(i)), nil
}

// convertInt16 converts a value to int16.
func convertInt16(value any) (reflect.Value, error) {
	i, err := convertToInt(value, 16)
	if err != nil {
		return reflect.Value{}, err
	}

	//nolint:gosec // Overflow checked by bitSize in convertToInt
	return reflect.ValueOf(int16(i)), nil
}

// convertInt32 converts a value to int32.
func convertInt32(value any) (reflect.Value, error) {
	i, err := convertToInt(value, 32)
	if err != nil {
		return reflect.Value{}, err
	}

	//nolint:gosec // Overflow checked by bitSize in convertToInt
	return reflect.ValueOf(int32(i)), nil
}

// convertInt64 converts a value to int64.
func convertInt64(value any) (reflect.Value, error) {
	i, err := convertToInt(value, 64)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(i), nil
}

// convertToInt converts a value to int64 with specified bit size for validation.
func convertToInt(value any, bitSize int) (int64, error) {
	dataVal := reflect.Indirect(reflect.ValueOf(value))
	kind := getKind(dataVal)

	//nolint:exhaustive // Only handling convertible types
	switch kind {
	case reflect.Int:
		return dataVal.Int(), nil
	case reflect.Uint:
		//nolint:gosec // Overflow is acceptable for weakly typed conversion
		return int64(dataVal.Uint()), nil
	case reflect.Float32:
		return int64(dataVal.Float()), nil
	case reflect.Bool:
		if dataVal.Bool() {
			return 1, nil
		}

		return 0, nil
	case reflect.String:
		return parseInt(dataVal.String(), bitSize)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// convertUint converts a value to uint.
func convertUint(value any) (reflect.Value, error) {
	u, err := convertToUint(value, 0)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(uint(u)), nil
}

// convertUint8 converts a value to uint8.
func convertUint8(value any) (reflect.Value, error) {
	u, err := convertToUint(value, 8)
	if err != nil {
		return reflect.Value{}, err
	}

	//nolint:gosec // Overflow checked by bitSize in convertToUint
	return reflect.ValueOf(uint8(u)), nil
}

// convertUint16 converts a value to uint16.
func convertUint16(value any) (reflect.Value, error) {
	u, err := convertToUint(value, 16)
	if err != nil {
		return reflect.Value{}, err
	}

	//nolint:gosec // Overflow checked by bitSize in convertToUint
	return reflect.ValueOf(uint16(u)), nil
}

// convertUint32 converts a value to uint32.
func convertUint32(value any) (reflect.Value, error) {
	u, err := convertToUint(value, 32)
	if err != nil {
		return reflect.Value{}, err
	}

	//nolint:gosec // Overflow checked by bitSize in convertToUint
	return reflect.ValueOf(uint32(u)), nil
}

// convertUint64 converts a value to uint64.
func convertUint64(value any) (reflect.Value, error) {
	u, err := convertToUint(value, 64)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(u), nil
}

// convertToUint converts a value to uint64 with specified bit size for validation.
func convertToUint(value any, bitSize int) (uint64, error) {
	dataVal := reflect.Indirect(reflect.ValueOf(value))
	kind := getKind(dataVal)

	//nolint:exhaustive // Only handling convertible types
	switch kind {
	case reflect.Int:
		return intToUint(dataVal.Int())
	case reflect.Uint:
		return dataVal.Uint(), nil
	case reflect.Float32:
		return floatToUint(dataVal.Float())
	case reflect.Bool:
		return boolToUint(dataVal.Bool()), nil
	case reflect.String:
		return parseUint(dataVal.String(), bitSize)
	default:
		return 0, fmt.Errorf("cannot convert %T to uint", value)
	}
}

// convertFloat32 converts a value to float32.
// Handles int, uint, float directly; bool converts to 0/1; parses string values.
func convertFloat32(value any) (reflect.Value, error) {
	f, err := convertToFloat(value, 32)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(float32(f)), nil
}

// convertFloat64 converts a value to float64.
// Handles int, uint, float directly; bool converts to 0/1; parses string values.
func convertFloat64(value any) (reflect.Value, error) {
	f, err := convertToFloat(value, 64)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(f), nil
}

// convertToFloat converts a value to float64 with specified bit size for validation.
func convertToFloat(value any, bitSize int) (float64, error) {
	dataVal := reflect.Indirect(reflect.ValueOf(value))
	kind := getKind(dataVal)

	//nolint:exhaustive // Only handling convertible types
	switch kind {
	case reflect.Int:
		return float64(dataVal.Int()), nil
	case reflect.Uint:
		return float64(dataVal.Uint()), nil
	case reflect.Float32:
		return dataVal.Float(), nil
	case reflect.Bool:
		if dataVal.Bool() {
			return 1, nil
		}

		return 0, nil
	case reflect.String:
		return parseFloat(dataVal.String(), bitSize)
	default:
		return 0, fmt.Errorf("cannot convert %T to float", value)
	}
}

func parseInt(s string, bitSize int) (int64, error) {
	if s == "" {
		return 0, nil
	}

	i, err := strconv.ParseInt(s, 0, bitSize)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as int: %w", s, err)
	}

	return i, nil
}

func parseUint(s string, bitSize int) (uint64, error) {
	if s == "" {
		return 0, nil
	}

	u, err := strconv.ParseUint(s, 0, bitSize)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as uint: %w", s, err)
	}

	return u, nil
}

func parseFloat(s string, bitSize int) (float64, error) {
	if s == "" {
		return 0, nil
	}

	f, err := strconv.ParseFloat(s, bitSize)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as float: %w", s, err)
	}

	return f, nil
}

func intToUint(i int64) (uint64, error) {
	if i < 0 {
		return 0, fmt.Errorf("cannot convert negative value %d to uint", i)
	}

	return uint64(i), nil
}

func floatToUint(f float64) (uint64, error) {
	if f < 0 {
		return 0, fmt.Errorf("cannot convert negative value %f to uint", f)
	}

	return uint64(f), nil
}

func boolToUint(b bool) uint64 {
	if b {
		return 1
	}

	return 0
}
