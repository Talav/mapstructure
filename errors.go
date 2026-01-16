package mapstructure

import (
	"fmt"
	"reflect"
)

// ConversionError represents a type conversion failure.
type ConversionError struct {
	FieldPath  string
	Value      any
	TargetType reflect.Type
	Cause      error
}

func (e *ConversionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: cannot convert %T to %v: %v",
			e.FieldPath, e.Value, e.TargetType, e.Cause)
	}

	return fmt.Sprintf("%s: cannot convert %T to %v",
		e.FieldPath, e.Value, e.TargetType)
}

func (e *ConversionError) Unwrap() error {
	return e.Cause
}

// NewConversionError creates a new ConversionError.
func NewConversionError(fieldPath string, value any, targetType reflect.Type, cause error) *ConversionError {
	if fieldPath == "" {
		fieldPath = "root"
	}

	return &ConversionError{
		FieldPath:  fieldPath,
		Value:      value,
		TargetType: targetType,
		Cause:      cause,
	}
}

// ValidationError represents a validation failure for the result pointer.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new ValidationError.
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}
