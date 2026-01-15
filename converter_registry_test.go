package mapstructure

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverterRegistry_NewWithConverters(t *testing.T) {
	typ := reflect.TypeOf(int(0))
	conv := func(value any) (reflect.Value, error) {
		return reflect.ValueOf(42), nil
	}

	registry := NewConverterRegistry(map[reflect.Type]Converter{
		typ: conv,
	})

	found, ok := registry.Find(typ)
	require.True(t, ok)
	assert.NotNil(t, found)
}

func TestConverterRegistry_NewWithNil(t *testing.T) {
	registry := NewConverterRegistry(nil)

	typ := reflect.TypeOf(int(0))
	_, ok := registry.Find(typ)
	assert.False(t, ok)
}

func TestConverterRegistry_DefaultRegistry_FindBuiltIn(t *testing.T) {
	registry := NewDefaultConverterRegistry(nil)

	typ := reflect.TypeOf(string(""))
	_, ok := registry.Find(typ)
	assert.True(t, ok)
}

func TestConverterRegistry_DefaultRegistry_WithAdditional(t *testing.T) {
	registry := NewDefaultConverterRegistry(map[reflect.Type]Converter{
		reflect.TypeOf(complex64(0)): func(value any) (reflect.Value, error) {
			return reflect.ValueOf(complex64(1 + 2i)), nil
		},
	})

	// Should find built-in converters
	_, ok := registry.Find(reflect.TypeOf(int(0)))
	assert.True(t, ok)

	// Should find custom converter
	found, ok := registry.Find(reflect.TypeOf(complex64(0)))
	require.True(t, ok)
	assert.NotNil(t, found)
}

func TestConverterRegistry_DefaultRegistry_OverrideBuiltIn(t *testing.T) {
	customConverter := func(value any) (reflect.Value, error) {
		return reflect.ValueOf(999), nil
	}

	registry := NewDefaultConverterRegistry(map[reflect.Type]Converter{
		reflect.TypeOf(int(0)): customConverter,
	})

	// Should find the overridden converter
	found, ok := registry.Find(reflect.TypeOf(int(0)))
	require.True(t, ok)

	// Verify it's the custom one
	result, err := found(42)
	require.NoError(t, err)
	//nolint:forcetypeassert // Test code - safe to assert
	assert.Equal(t, 999, result.Interface().(int))
}
