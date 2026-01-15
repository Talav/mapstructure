package mapstructure

import (
	"fmt"
	"reflect"
	"strconv"
)

// convertBool converts a value to bool.
// Handles bool, int, uint, float directly; parses string values.
func convertBool(value any) (reflect.Value, error) {
	dataVal := reflect.Indirect(reflect.ValueOf(value))
	kind := getKind(dataVal)

	//nolint:exhaustive // Only handling convertible types
	switch kind {
	case reflect.Bool:
		return reflect.ValueOf(dataVal.Bool()), nil
	case reflect.Int:
		return reflect.ValueOf(dataVal.Int() != 0), nil
	case reflect.Uint:
		return reflect.ValueOf(dataVal.Uint() != 0), nil
	case reflect.Float32:
		return reflect.ValueOf(dataVal.Float() != 0), nil
	case reflect.String:
		s := dataVal.String()
		if s == "" {
			return reflect.ValueOf(false), nil
		}
		if b, err := strconv.ParseBool(s); err == nil {
			return reflect.ValueOf(b), nil
		}

		return reflect.Value{}, fmt.Errorf("cannot parse %q as bool", s)
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// convertString converts a value to string.
// Handles string, bool, int, uint, float, and []byte directly.
func convertString(value any) (reflect.Value, error) {
	dataVal := reflect.Indirect(reflect.ValueOf(value))
	kind := getKind(dataVal)

	//nolint:exhaustive // Only handling convertible types
	switch kind {
	case reflect.String:
		return reflect.ValueOf(dataVal.String()), nil
	case reflect.Bool:
		if dataVal.Bool() {
			return reflect.ValueOf("1"), nil
		}

		return reflect.ValueOf("0"), nil
	case reflect.Int:
		return reflect.ValueOf(strconv.FormatInt(dataVal.Int(), 10)), nil
	case reflect.Uint:
		return reflect.ValueOf(strconv.FormatUint(dataVal.Uint(), 10)), nil
	case reflect.Float32:
		return reflect.ValueOf(strconv.FormatFloat(dataVal.Float(), 'f', -1, 64)), nil
	case reflect.Slice:
		// Handle []byte
		if dataVal.Type().Elem().Kind() == reflect.Uint8 {
			//nolint:forcetypeassert // Type checked above
			bytes := dataVal.Interface().([]byte)

			return reflect.ValueOf(string(bytes)), nil
		}

		return reflect.Value{}, fmt.Errorf("cannot convert %T to string", value)
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to string", value)
	}
}

// getKind normalizes reflect.Kind to base types.
func getKind(val reflect.Value) reflect.Kind {
	kind := val.Kind()

	//nolint:exhaustive // Only normalizing numeric types
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.Int
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.Uint
	case reflect.Float32, reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
}
