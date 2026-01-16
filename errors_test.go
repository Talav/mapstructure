package mapstructure

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConversionError(t *testing.T) {
	t.Run("error message with cause", func(t *testing.T) {
		err := NewConversionError("user.age", "invalid", reflect.TypeOf(0), errors.New("parse error"))
		assert.Equal(t, "user.age: cannot convert string to int: parse error", err.Error())
	})

	t.Run("error message without cause", func(t *testing.T) {
		err := NewConversionError("user.name", 123, reflect.TypeOf(""), nil)
		assert.Equal(t, "user.name: cannot convert int to string", err.Error())
	})

	t.Run("empty field path defaults to root", func(t *testing.T) {
		err := NewConversionError("", true, reflect.TypeOf(0), nil)
		assert.Equal(t, "root", err.FieldPath)
	})

	t.Run("unwrap returns cause", func(t *testing.T) {
		cause := errors.New("underlying")
		err := NewConversionError("field", "val", reflect.TypeOf(0), cause)
		unwrapped := err.Unwrap()
		require.NotNil(t, unwrapped)
		assert.Equal(t, cause.Error(), unwrapped.Error())
	})

	t.Run("errors.Is integration", func(t *testing.T) {
		cause := errors.New("root cause")
		err := NewConversionError("field", "value", reflect.TypeOf(0), cause)
		assert.True(t, errors.Is(err, cause))
	})
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("result must be a non-nil pointer")
	assert.Equal(t, "result must be a non-nil pointer", err.Error())
}

// Compile-time interface checks.
var (
	_ error = (*ConversionError)(nil)
	_ error = (*ValidationError)(nil)
)
