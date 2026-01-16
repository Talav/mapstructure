package mapstructure

import (
	"fmt"
	"reflect"
)

var defaultUnmarshaler = &Unmarshaler{
	fieldCache: NewDefaultStructMetadataCache(),
	converters: NewDefaultConverterRegistry(),
}

// Unmarshal transforms map[string]any into a Go struct pointed to by result.
// result must be a pointer to the target type.
// This is a convenience function that uses a shared default unmarshaler.
func Unmarshal(data map[string]any, result any) error {
	return defaultUnmarshaler.Unmarshal(data, result)
}

// Unmarshaler handles unmarshaling of maps to Go structs.
type Unmarshaler struct {
	fieldCache *StructMetadataCache
	converters *ConverterRegistry
}

// NewUnmarshaler creates a new unmarshaler with explicit dependencies.
// For custom configurations, construct the cache and converters separately:
//
//	// With custom tag
//	cache := NewStructMetadataCache("json", "default")
//	converters := NewDefaultConverterRegistry()
//	u := NewUnmarshaler(cache, converters)
//
//	// With custom converters
//	cache := NewDefaultStructMetadataCache()
//	converters := NewDefaultConverterRegistry(customConverters)
//	u := NewUnmarshaler(cache, converters)
func NewUnmarshaler(fieldCache *StructMetadataCache, converters *ConverterRegistry) *Unmarshaler {
	return &Unmarshaler{
		fieldCache: fieldCache,
		converters: converters,
	}
}

// NewDefaultUnmarshaler creates a new unmarshaler with default settings.
// Uses "schema" tags for field mapping and "default" tags for default values.
func NewDefaultUnmarshaler() *Unmarshaler {
	return NewUnmarshaler(NewDefaultStructMetadataCache(), NewDefaultConverterRegistry())
}

// Unmarshal transforms map[string]any into a Go struct pointed to by result.
// result must be a pointer to the target type.
func (u *Unmarshaler) Unmarshal(data map[string]any, result any) error {
	rv, err := validateResultPointer(result)
	if err != nil {
		return err
	}

	return u.unmarshalValue(data, rv, "")
}

// unmarshalValue recursively unmarshals a value into the reflect.Value.
func (u *Unmarshaler) unmarshalValue(data any, rv reflect.Value, fieldPath string) error {
	if !rv.CanSet() {
		return nil
	}

	kind := rv.Kind()
	typ := rv.Type()

	// Direct assignment if types are compatible
	if data != nil {
		dataType := reflect.TypeOf(data)
		if dataType.AssignableTo(typ) {
			rv.Set(reflect.ValueOf(data))

			return nil
		}
	}

	// Try converter for the target type
	if conv, ok := u.converters.Find(typ); ok {
		converted, err := conv(data)
		if err != nil {
			return NewConversionError(fieldPath, data, typ, err)
		}
		rv.Set(converted)

		return nil
	}

	//nolint:exhaustive // Unsupported types are handled in default case with error
	switch kind {
	case reflect.Ptr:
		return u.unmarshalPtr(data, rv, fieldPath)
	case reflect.Slice:
		return u.unmarshalSlice(data, rv, fieldPath)
	case reflect.Struct:
		return u.unmarshalStruct(data, rv, fieldPath)
	default:
		return fmt.Errorf("%s: no converter registered for type %v", fieldPath, typ)
	}
}

// unmarshalPtr unmarshals a pointer value.
func (u *Unmarshaler) unmarshalPtr(data any, rv reflect.Value, fieldPath string) error {
	// If data is nil or missing, set pointer to nil
	if data == nil {
		rv.Set(reflect.Zero(rv.Type()))

		return nil
	}

	// If pointer is nil, allocate new instance
	if rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	// Recursively unmarshal the pointed-to type
	return u.unmarshalValue(data, rv.Elem(), fieldPath)
}

// unmarshalSlice unmarshals a slice value.
func (u *Unmarshaler) unmarshalSlice(data any, rv reflect.Value, fieldPath string) error {
	// nil is acceptable for slices
	if data == nil {
		rv.Set(reflect.Zero(rv.Type()))

		return nil
	}

	// Use reflection to handle any slice type ([]any, []byte, []int, etc.)
	dataVal := reflect.ValueOf(data)
	if dataVal.Kind() != reflect.Slice && dataVal.Kind() != reflect.Array {
		return NewConversionError(fieldPath, data, rv.Type(), nil)
	}

	dataLen := dataVal.Len()
	if dataLen == 0 {
		rv.Set(reflect.MakeSlice(rv.Type(), 0, 0))

		return nil
	}

	return u.unmarshalSliceElements(dataVal, rv, fieldPath, dataLen)
}

// unmarshalSliceElements handles the actual slice element unmarshaling with fast paths.
func (u *Unmarshaler) unmarshalSliceElements(dataVal, rv reflect.Value, fieldPath string, dataLen int) error {
	// Pre-allocate slice with appropriate capacity
	slice := reflect.MakeSlice(rv.Type(), dataLen, dataLen)
	sliceElemType := slice.Type().Elem()

	// Fast path 1: direct assignment for fully compatible types
	if dataVal.Type().AssignableTo(slice.Type()) {
		rv.Set(dataVal)

		return nil
	}

	// Fast path 2: direct copy for same element type
	if dataVal.Type().Elem() == sliceElemType {
		reflect.Copy(slice, dataVal)
		rv.Set(slice)

		return nil
	}

	// Fast path 3: direct element assignment for interface targets
	if sliceElemType.Kind() == reflect.Interface {
		for i := range dataLen {
			slice.Index(i).Set(dataVal.Index(i))
		}
		rv.Set(slice)

		return nil
	}

	// Regular conversion path: element-by-element with converters
	for i := range dataLen {
		elemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
		if err := u.unmarshalValue(dataVal.Index(i).Interface(), slice.Index(i), elemPath); err != nil {
			return err
		}
	}

	rv.Set(slice)

	return nil
}

// unmarshalStruct unmarshals a struct value using cached field metadata.
func (u *Unmarshaler) unmarshalStruct(data any, rv reflect.Value, fieldPath string) error {
	// Expect map[string]any for struct data
	dataMap, ok := data.(map[string]any)
	if !ok {
		return NewConversionError(fieldPath, data, rv.Type(), nil)
	}

	// Get cached fields
	typ := rv.Type()
	metadata := u.fieldCache.GetMetadata(typ)

	// Process each cached field
	for _, field := range metadata.Fields {
		fieldValue := rv.Field(field.Index)

		// Handle embedded structs
		if field.Embedded {
			if err := u.unmarshalEmbeddedField(dataMap, fieldValue, field, fieldPath); err != nil {
				return err
			}

			continue
		}

		// Get value from map, fall back to default if not present
		value, exists := dataMap[field.MapKey]
		if !exists {
			if field.Default == nil {
				continue
			}

			value = *field.Default
		}

		// Unmarshal the field value (handles converters and built-in conversion)
		fullPath := buildFieldPath(fieldPath, field.MapKey)
		if err := u.unmarshalValue(value, fieldValue, fullPath); err != nil {
			return fmt.Errorf("%s: %w", fullPath, err)
		}
	}

	return nil
}

// unmarshalEmbeddedField handles unmarshaling of embedded struct fields.
func (u *Unmarshaler) unmarshalEmbeddedField(dataMap map[string]any, fieldValue reflect.Value, field FieldMetadata, fieldPath string) error {
	if field.Type.Kind() != reflect.Struct {
		return nil
	}

	// Check if there's a nested map with the field name (named embedded)
	if nestedMap, exists := dataMap[field.StructFieldName]; exists {
		if nestedData, ok := nestedMap.(map[string]any); ok {
			// Named embedded: unmarshal from nested map
			return u.unmarshalValue(nestedData, fieldValue, fieldPath)
		}
	}

	// Anonymous embedded: pass entire data map (promoted fields)
	return u.unmarshalValue(dataMap, fieldValue, fieldPath)
}

// validateResultPointer validates that result is a non-nil pointer and returns its element.
func validateResultPointer(result any) (reflect.Value, error) {
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Ptr {
		return reflect.Value{}, NewValidationError("result must be a pointer")
	}

	if rv.IsNil() {
		return reflect.Value{}, NewValidationError("result pointer is nil")
	}

	return rv.Elem(), nil
}

// buildFieldPath builds a field path for error messages.
func buildFieldPath(base, field string) string {
	if base == "" {
		return field
	}

	return base + "." + field
}
