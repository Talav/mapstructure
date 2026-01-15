package mapstructure

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
)

// convertBytes converts a value to []byte.
// Handles []byte, string, []any, io.ReadCloser, and io.Reader.
func convertBytes(value any) (reflect.Value, error) {
	if value == nil {
		return reflect.ValueOf([]byte(nil)), nil
	}

	switch v := value.(type) {
	case []byte:
		return reflect.ValueOf(v), nil
	case string:
		return reflect.ValueOf([]byte(v)), nil
	case []any:
		return convertAnySliceToBytes(v)
	case io.ReadCloser:
		defer func() { _ = v.Close() }()

		b, err := io.ReadAll(v)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to read: %w", err)
		}

		return reflect.ValueOf(b), nil
	case io.Reader:
		b, err := io.ReadAll(v)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to read: %w", err)
		}

		return reflect.ValueOf(b), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to []byte", value)
}

// convertAnySliceToBytes converts []any to []byte by converting each element.
func convertAnySliceToBytes(slice []any) (reflect.Value, error) {
	result := make([]byte, len(slice))
	for i, elem := range slice {
		b, err := convertToUint(elem, 8)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("element [%d]: %w", i, err)
		}
		result[i] = byte(b)
	}

	return reflect.ValueOf(result), nil
}

// convertReadCloser converts a value to io.ReadCloser.
// Handles io.ReadCloser (passthrough), []byte, and string.
func convertReadCloser(value any) (reflect.Value, error) {
	if value == nil {
		return reflect.ValueOf((*io.ReadCloser)(nil)), nil
	}

	// Direct type checks
	switch v := value.(type) {
	case io.ReadCloser:
		return reflect.ValueOf(v), nil
	case io.Reader:
		return reflect.ValueOf(io.NopCloser(v)), nil
	case []byte:
		return reflect.ValueOf(io.NopCloser(bytes.NewReader(v))), nil
	case string:
		return reflect.ValueOf(io.NopCloser(bytes.NewReader([]byte(v)))), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to io.ReadCloser", value)
}
