package mapstructure

import "reflect"

// Converter converts a value to a reflect.Value of a specific type.
type Converter func(value any) (reflect.Value, error)
