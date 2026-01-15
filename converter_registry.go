package mapstructure

import (
	"io"
	"maps"
	"reflect"
)

// ConverterRegistry manages type converters.
// Immutable after construction, safe for concurrent reads.
type ConverterRegistry struct {
	converters map[reflect.Type]Converter
}

// NewConverterRegistry creates a registry with the given converters.
// If converters is nil, an empty registry is created.
func NewConverterRegistry(converters map[reflect.Type]Converter) *ConverterRegistry {
	if converters == nil {
		converters = make(map[reflect.Type]Converter)
	}

	return &ConverterRegistry{
		converters: converters,
	}
}

// NewDefaultConverterRegistry creates a registry with standard type converters.
// Additional converters can be provided to extend or override defaults.
func NewDefaultConverterRegistry(additional map[reflect.Type]Converter) *ConverterRegistry {
	converters := map[reflect.Type]Converter{
		reflect.TypeOf(string("")):                   convertString,
		reflect.TypeOf(bool(false)):                  convertBool,
		reflect.TypeOf(int(0)):                       convertInt,
		reflect.TypeOf(int8(0)):                      convertInt8,
		reflect.TypeOf(int16(0)):                     convertInt16,
		reflect.TypeOf(int32(0)):                     convertInt32,
		reflect.TypeOf(int64(0)):                     convertInt64,
		reflect.TypeOf(uint(0)):                      convertUint,
		reflect.TypeOf(uint8(0)):                     convertUint8,
		reflect.TypeOf(uint16(0)):                    convertUint16,
		reflect.TypeOf(uint32(0)):                    convertUint32,
		reflect.TypeOf(uint64(0)):                    convertUint64,
		reflect.TypeOf(float32(0)):                   convertFloat32,
		reflect.TypeOf(float64(0)):                   convertFloat64,
		reflect.TypeOf([]byte(nil)):                  convertBytes,
		reflect.TypeOf((*io.ReadCloser)(nil)).Elem(): convertReadCloser,
	}

	// Merge additional converters (allows override)
	maps.Copy(converters, additional)

	return &ConverterRegistry{
		converters: converters,
	}
}

// Find finds a converter for the given type.
// Lock-free read, safe for concurrent use.
func (r *ConverterRegistry) Find(typ reflect.Type) (Converter, bool) {
	conv, ok := r.converters[typ]

	return conv, ok
}
