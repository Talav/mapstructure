package mapstructure

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConverterRegistry(t *testing.T) {
	t.Run("with custom converters", func(t *testing.T) {
		customConverter := func(value any) (reflect.Value, error) {
			return reflect.ValueOf(42), nil
		}

		registry := NewConverterRegistry(map[reflect.Type]Converter{
			reflect.TypeOf(int(0)): customConverter,
		})

		found, ok := registry.Find(reflect.TypeOf(int(0)))
		require.True(t, ok)
		assert.NotNil(t, found)
	})

	t.Run("with nil map", func(t *testing.T) {
		registry := NewConverterRegistry(nil)

		_, ok := registry.Find(reflect.TypeOf(int(0)))
		assert.False(t, ok, "empty registry should not find any converters")
	})
}

func TestNewDefaultConverterRegistry(t *testing.T) {
	t.Run("with no arguments", func(t *testing.T) {
		registry := NewDefaultConverterRegistry()

		// Should have all built-in converters
		builtInTypes := []reflect.Type{
			reflect.TypeOf(int(0)),
			reflect.TypeOf(string("")),
			reflect.TypeOf(bool(false)),
			reflect.TypeOf(float64(0)),
			reflect.TypeOf([]byte(nil)),
		}

		for _, typ := range builtInTypes {
			converter, ok := registry.Find(typ)
			assert.True(t, ok, "should find built-in converter for %v", typ)
			assert.NotNil(t, converter)
		}
	})

	t.Run("with additional converters", func(t *testing.T) {
		customConverter := func(value any) (reflect.Value, error) {
			return reflect.ValueOf(complex64(1 + 2i)), nil
		}
		overrideConverter := func(value any) (reflect.Value, error) {
			return reflect.ValueOf(999), nil
		}

		registry := NewDefaultConverterRegistry(map[reflect.Type]Converter{
			reflect.TypeOf(complex64(0)): customConverter,   // Add new type
			reflect.TypeOf(int(0)):       overrideConverter, // Override built-in
		})

		// Should still have other built-in converters
		_, ok := registry.Find(reflect.TypeOf(string("")))
		assert.True(t, ok, "should preserve other built-in converters")

		// Should have new custom converter
		found, ok := registry.Find(reflect.TypeOf(complex64(0)))
		require.True(t, ok, "should find custom converter")
		assert.NotNil(t, found)

		// Should use overridden converter
		found, ok = registry.Find(reflect.TypeOf(int(0)))
		require.True(t, ok)
		result, err := found(42)
		require.NoError(t, err)
		//nolint:forcetypeassert // Test code - safe to assert
		assert.Equal(t, 999, result.Interface().(int), "should use overridden converter")
	})

	t.Run("with multiple maps merges in order", func(t *testing.T) {
		conv1 := func(value any) (reflect.Value, error) { return reflect.ValueOf(111), nil }
		conv2 := func(value any) (reflect.Value, error) { return reflect.ValueOf(222), nil }

		registry := NewDefaultConverterRegistry(
			map[reflect.Type]Converter{reflect.TypeOf(int(0)): conv1},
			map[reflect.Type]Converter{reflect.TypeOf(int(0)): conv2}, // Later should win
		)

		found, ok := registry.Find(reflect.TypeOf(int(0)))
		require.True(t, ok)
		result, err := found(0)
		require.NoError(t, err)
		//nolint:forcetypeassert // Test code - safe to assert
		assert.Equal(t, 222, result.Interface().(int), "later map should override earlier")
	})
}
