package mapstructure

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testUnmarshaler creates an unmarshaler that uses field names directly (no tags).
func testUnmarshaler() *Unmarshaler {
	// "-" means use field names directly
	cache := NewStructMetadataCache("-", "")

	return NewUnmarshaler(cache, NewDefaultConverterRegistry())
}

func TestUnmarshaler_Unmarshal_Slices(t *testing.T) {
	type SliceInt struct {
		Values []int
	}
	type SliceByte struct {
		Data []byte
	}
	type SliceString struct {
		Names []string
	}

	tests := []struct {
		name     string
		data     map[string]any
		target   any
		expected any
	}{
		{
			name:     "slice of any to int",
			data:     map[string]any{"Values": []any{1, 2, 3}},
			target:   &SliceInt{},
			expected: &SliceInt{Values: []int{1, 2, 3}},
		},
		{
			name:     "slice of int to int",
			data:     map[string]any{"Values": []int{10, 20, 30}},
			target:   &SliceInt{},
			expected: &SliceInt{Values: []int{10, 20, 30}},
		},
		{
			name:     "slice of bytes direct",
			data:     map[string]any{"Data": []byte{1, 2, 3, 4, 5}},
			target:   &SliceByte{},
			expected: &SliceByte{Data: []byte{1, 2, 3, 4, 5}},
		},
		{
			name:     "slice of any to bytes",
			data:     map[string]any{"Data": []any{1, 2, 3, 4, 5}},
			target:   &SliceByte{},
			expected: &SliceByte{Data: []byte{1, 2, 3, 4, 5}},
		},
		{
			name:     "slice of strings",
			data:     map[string]any{"Names": []string{"alice", "bob", "charlie"}},
			target:   &SliceString{},
			expected: &SliceString{Names: []string{"alice", "bob", "charlie"}},
		},
		{
			name:     "nil slice",
			data:     map[string]any{"Data": nil},
			target:   &SliceByte{},
			expected: &SliceByte{Data: nil},
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestUnmarshaler_Unmarshal_BasicTypes(t *testing.T) {
	type Basic struct {
		Name   string
		Age    int
		Score  float64
		Active bool
	}

	tests := []struct {
		name     string
		data     map[string]any
		expected Basic
	}{
		{
			name: "all fields",
			data: map[string]any{
				"Name":   "test",
				"Age":    25,
				"Score":  95.5,
				"Active": true,
			},
			expected: Basic{Name: "test", Age: 25, Score: 95.5, Active: true},
		},
		{
			name: "missing field uses zero value",
			data: map[string]any{
				"Name": "test",
			},
			expected: Basic{Name: "test", Age: 0, Score: 0, Active: false},
		},
		{
			name:     "empty map",
			data:     map[string]any{},
			expected: Basic{},
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Basic
			err := u.Unmarshal(tt.data, &result)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnmarshaler_Unmarshal_NestedStruct(t *testing.T) {
	type Inner struct {
		Value int
	}
	type Outer struct {
		Inner Inner
	}
	type DeepNested struct {
		Level1 struct {
			Level2 struct {
				Value string
			}
		}
	}

	tests := []struct {
		name     string
		data     map[string]any
		target   any
		expected any
	}{
		{
			name: "single level nested",
			data: map[string]any{
				"Inner": map[string]any{"Value": 42},
			},
			target:   &Outer{},
			expected: &Outer{Inner: Inner{Value: 42}},
		},
		{
			name: "deep nested",
			data: map[string]any{
				"Level1": map[string]any{
					"Level2": map[string]any{
						"Value": "deep",
					},
				},
			},
			target:   &DeepNested{},
			expected: &DeepNested{Level1: struct{ Level2 struct{ Value string } }{Level2: struct{ Value string }{Value: "deep"}}},
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestUnmarshaler_Unmarshal_Pointers(t *testing.T) {
	type WithPointer struct {
		Value *int
	}
	type WithPointerString struct {
		Name *string
	}

	intVal := 42
	strVal := "test"

	tests := []struct {
		name     string
		data     map[string]any
		target   any
		expected any
	}{
		{
			name:     "non-nil pointer",
			data:     map[string]any{"Value": 42},
			target:   &WithPointer{},
			expected: &WithPointer{Value: &intVal},
		},
		{
			name:     "nil pointer",
			data:     map[string]any{"Value": nil},
			target:   &WithPointer{},
			expected: &WithPointer{Value: nil},
		},
		{
			name:     "string pointer",
			data:     map[string]any{"Name": "test"},
			target:   &WithPointerString{},
			expected: &WithPointerString{Name: &strVal},
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestUnmarshaler_Unmarshal_Errors(t *testing.T) {
	type Target struct {
		Data []int
		Name string
	}

	tests := []struct {
		name        string
		data        map[string]any
		target      any
		errContains string
	}{
		{
			name:        "invalid slice data",
			data:        map[string]any{"Data": "not a slice"},
			target:      &Target{},
			errContains: "cannot convert",
		},
		{
			name:        "non-pointer result",
			data:        map[string]any{"Name": "test"},
			target:      Target{},
			errContains: "must be a pointer",
		},
		{
			name:        "nil pointer result",
			data:        map[string]any{"Name": "test"},
			target:      (*Target)(nil),
			errContains: "nil",
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

//nolint:forcetypeassert,thelper // Test code - type assertions are expected to succeed
func TestUnmarshaler_Unmarshal_DirectAssignment(t *testing.T) {
	type Inner struct {
		Value int
		Name  string
	}

	type WithReader struct {
		Reader io.Reader
	}
	type WithReadCloser struct {
		Body io.ReadCloser
	}
	type WithWriter struct {
		Writer io.Writer
	}
	type WithBytes struct {
		Data []byte
	}
	type WithTime struct {
		CreatedAt time.Time
	}
	type WithInner struct {
		Inner Inner
	}
	type WithInnerPtr struct {
		Inner *Inner
	}
	type WithInnerSlice struct {
		Items []Inner
	}
	type WithMap struct {
		Metadata map[string]string
	}

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		data     map[string]any
		target   any
		validate func(t *testing.T, target any)
	}{
		{
			name:   "[]byte direct assignment",
			data:   map[string]any{"Data": []byte{0x01, 0x02, 0x03}},
			target: &WithBytes{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithBytes)
				assert.Equal(t, []byte{0x01, 0x02, 0x03}, r.Data)
			},
		},
		{
			name:   "[]byte empty",
			data:   map[string]any{"Data": []byte{}},
			target: &WithBytes{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithBytes)
				assert.Equal(t, []byte{}, r.Data)
			},
		},
		{
			name:   "io.Reader direct assignment",
			data:   map[string]any{"Reader": bytes.NewReader([]byte("hello"))},
			target: &WithReader{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithReader)
				require.NotNil(t, r.Reader)
				content, err := io.ReadAll(r.Reader)
				require.NoError(t, err)
				assert.Equal(t, []byte("hello"), content)
			},
		},
		{
			name:   "io.ReadCloser direct assignment",
			data:   map[string]any{"Body": io.NopCloser(bytes.NewReader([]byte("body")))},
			target: &WithReadCloser{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithReadCloser)
				require.NotNil(t, r.Body)
				content, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.Equal(t, []byte("body"), content)
			},
		},
		{
			name:   "io.Writer interface satisfaction",
			data:   map[string]any{"Writer": &bytes.Buffer{}},
			target: &WithWriter{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithWriter)
				require.NotNil(t, r.Writer)
				_, err := r.Writer.Write([]byte("test"))
				assert.NoError(t, err)
			},
		},
		{
			name:   "custom struct direct assignment",
			data:   map[string]any{"Inner": Inner{Value: 42, Name: "test"}},
			target: &WithInner{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithInner)
				assert.Equal(t, Inner{Value: 42, Name: "test"}, r.Inner)
			},
		},
		{
			name:   "custom struct pointer direct assignment",
			data:   map[string]any{"Inner": &Inner{Value: 99, Name: "ptr"}},
			target: &WithInnerPtr{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithInnerPtr)
				require.NotNil(t, r.Inner)
				assert.Equal(t, 99, r.Inner.Value)
				assert.Equal(t, "ptr", r.Inner.Name)
			},
		},
		{
			name:   "slice of custom structs",
			data:   map[string]any{"Items": []Inner{{Value: 1, Name: "a"}, {Value: 2, Name: "b"}}},
			target: &WithInnerSlice{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithInnerSlice)
				require.Len(t, r.Items, 2)
				assert.Equal(t, "a", r.Items[0].Name)
				assert.Equal(t, "b", r.Items[1].Name)
			},
		},
		{
			name:   "map[string]string direct assignment",
			data:   map[string]any{"Metadata": map[string]string{"key": "value"}},
			target: &WithMap{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithMap)
				assert.Equal(t, "value", r.Metadata["key"])
			},
		},
		{
			name:   "time.Time direct assignment",
			data:   map[string]any{"CreatedAt": now},
			target: &WithTime{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithTime)
				assert.Equal(t, now, r.CreatedAt)
			},
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.NoError(t, err)
			tt.validate(t, tt.target)
		})
	}
}

func TestUnmarshaler_Unmarshal_DirectAssignment_FallbackToConverter(t *testing.T) {
	type Target struct {
		Count int
	}

	tests := []struct {
		name     string
		data     map[string]any
		expected int
	}{
		{
			name:     "string to int uses converter",
			data:     map[string]any{"Count": "42"},
			expected: 42,
		},
		{
			name:     "int64 to int uses converter",
			data:     map[string]any{"Count": int64(42)},
			expected: 42,
		},
		{
			name:     "float64 to int uses converter",
			data:     map[string]any{"Count": float64(42.0)},
			expected: 42,
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Target
			err := u.Unmarshal(tt.data, &result)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Count)
		})
	}
}

func TestUnmarshaler_Unmarshal_DefaultValues(t *testing.T) {
	type WithDefaults struct {
		Name   string  `schema:"name" default:"anonymous"`
		Count  int     `schema:"count" default:"10"`
		Score  float64 `schema:"score" default:"99.5"`
		Active bool    `schema:"active" default:"true"`
	}

	type WithPointerDefault struct {
		Value *int `schema:"value" default:"42"`
	}

	intPtr := func(v int) *int { return &v }

	type Mixed struct {
		Required string `schema:"required"`
		Optional string `schema:"optional" default:"default_value"`
	}

	tests := []struct {
		name     string
		data     map[string]any
		target   any
		expected any
	}{
		{
			name:   "all defaults applied",
			data:   map[string]any{},
			target: &WithDefaults{},
			expected: &WithDefaults{
				Name:   "anonymous",
				Count:  10,
				Score:  99.5,
				Active: true,
			},
		},
		{
			name:   "explicit values override defaults",
			data:   map[string]any{"name": "custom", "count": 99},
			target: &WithDefaults{},
			expected: &WithDefaults{
				Name:   "custom",
				Count:  99,
				Score:  99.5,
				Active: true,
			},
		},
		{
			name:     "pointer default",
			data:     map[string]any{},
			target:   &WithPointerDefault{},
			expected: &WithPointerDefault{Value: intPtr(42)},
		},
		{
			name:   "mixed required and optional",
			data:   map[string]any{"required": "provided"},
			target: &Mixed{},
			expected: &Mixed{
				Required: "provided",
				Optional: "default_value",
			},
		},
	}

	u := NewDefaultUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestUnmarshaler_Unmarshal_DefaultValues_CustomConverter(t *testing.T) {
	type Status int

	const (
		StatusPending Status = iota
		StatusActive
		StatusClosed
	)

	type WithCustomDefault struct {
		Status Status `schema:"status" default:"active"`
	}

	// Custom converter for Status type
	statusConverter := func(v any) (reflect.Value, error) {
		s, ok := v.(string)
		if !ok {
			return reflect.Value{}, nil
		}

		switch s {
		case "pending":
			return reflect.ValueOf(StatusPending), nil
		case "active":
			return reflect.ValueOf(StatusActive), nil
		case "closed":
			return reflect.ValueOf(StatusClosed), nil
		default:
			return reflect.Value{}, nil
		}
	}

	converters := map[reflect.Type]Converter{
		reflect.TypeOf(Status(0)): statusConverter,
	}

	// Create unmarshaler with custom converters
	cache := NewStructMetadataCache("schema", "")
	convertersRegistry := NewDefaultConverterRegistry(converters)
	u := NewUnmarshaler(cache, convertersRegistry)

	var result WithCustomDefault
	err := u.Unmarshal(map[string]any{}, &result)

	require.NoError(t, err)
	assert.Equal(t, StatusActive, result.Status)
}

func TestUnmarshal_ConvenienceAPI(t *testing.T) {
	type Person struct {
		Name   string `schema:"name"`
		Age    int    `schema:"age"`
		Active bool   `schema:"active"`
	}

	tests := []struct {
		name     string
		data     map[string]any
		expected Person
	}{
		{
			name: "basic struct",
			data: map[string]any{
				"name":   "Alice",
				"age":    30,
				"active": true,
			},
			expected: Person{Name: "Alice", Age: 30, Active: true},
		},
		{
			name: "with type conversion",
			data: map[string]any{
				"name":   "Bob",
				"age":    "25", // string to int
				"active": 1,    // int to bool
			},
			expected: Person{Name: "Bob", Age: 25, Active: true},
		},
		{
			name: "partial data",
			data: map[string]any{
				"name": "Charlie",
			},
			expected: Person{Name: "Charlie", Age: 0, Active: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Person
			err := Unmarshal(tt.data, &result)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnmarshal_ConvenienceAPI_Errors(t *testing.T) {
	type Target struct {
		Name string `schema:"name"`
	}

	tests := []struct {
		name        string
		data        map[string]any
		target      any
		errContains string
	}{
		{
			name:        "non-pointer",
			data:        map[string]any{"name": "test"},
			target:      Target{},
			errContains: "must be a pointer",
		},
		{
			name:        "nil pointer",
			data:        map[string]any{"name": "test"},
			target:      (*Target)(nil),
			errContains: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal(tt.data, tt.target)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestUnmarshaler_Unmarshal_EmbeddedStructs(t *testing.T) {
	type Timestamps struct {
		CreatedAt string `schema:"created_at"`
		UpdatedAt string `schema:"updated_at"`
	}

	type Metadata struct {
		Version string `schema:"version"`
		Author  string `schema:"author"`
	}

	type User struct {
		Timestamps        // Embedded - anonymous
		Name       string `schema:"name"`
		Email      string `schema:"email"`
	}

	type Document struct {
		Metadata        // Embedded - anonymous
		Title    string `schema:"title"`
		Content  string `schema:"content"`
	}

	type MultiEmbedded struct {
		Timestamps        // First embedded
		Metadata          // Second embedded
		Name       string `schema:"name"`
	}

	t.Run("promoted fields", func(t *testing.T) {
		data := map[string]any{
			"name":       "Alice",
			"email":      "alice@example.com",
			"created_at": "2024-01-01",
			"updated_at": "2024-01-02",
		}

		var result User
		u := NewDefaultUnmarshaler()
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Equal(t, "Alice", result.Name)
		assert.Equal(t, "alice@example.com", result.Email)
		assert.Equal(t, "2024-01-01", result.CreatedAt)
		assert.Equal(t, "2024-01-02", result.UpdatedAt)
	})

	t.Run("named embedded access", func(t *testing.T) {
		data := map[string]any{
			"name":  "Alice",
			"email": "alice@example.com",
			"Timestamps": map[string]any{
				"created_at": "2024-01-01",
				"updated_at": "2024-01-02",
			},
		}

		var result User
		u := NewDefaultUnmarshaler()
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Equal(t, "Alice", result.Name)
		assert.Equal(t, "2024-01-01", result.CreatedAt)
		assert.Equal(t, "2024-01-02", result.UpdatedAt)
	})

	t.Run("multiple embedded structs", func(t *testing.T) {
		data := map[string]any{
			"name":       "Document 1",
			"created_at": "2024-01-01",
			"updated_at": "2024-01-02",
			"version":    "1.0",
			"author":     "Alice",
		}

		var result MultiEmbedded
		u := NewDefaultUnmarshaler()
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Equal(t, "Document 1", result.Name)
		assert.Equal(t, "2024-01-01", result.CreatedAt)
		assert.Equal(t, "2024-01-02", result.UpdatedAt)
		assert.Equal(t, "1.0", result.Version)
		assert.Equal(t, "Alice", result.Author)
	})

	t.Run("named access takes precedence", func(t *testing.T) {
		// When both named map exists, it uses named access
		data := map[string]any{
			"title":   "My Document",
			"content": "Some content",
			"Metadata": map[string]any{
				"version": "2.0",
				"author":  "Bob",
			},
		}

		var result Document
		u := NewDefaultUnmarshaler()
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Equal(t, "My Document", result.Title)
		assert.Equal(t, "Some content", result.Content)
		assert.Equal(t, "2.0", result.Version)
		assert.Equal(t, "Bob", result.Author)
	})

	t.Run("empty embedded struct", func(t *testing.T) {
		data := map[string]any{
			"name":  "Test",
			"email": "test@example.com",
		}

		var result User
		u := NewDefaultUnmarshaler()
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Equal(t, "Test", result.Name)
		assert.Equal(t, "", result.CreatedAt)
		assert.Equal(t, "", result.UpdatedAt)
	})
}

func TestUnmarshaler_Unmarshal_EmbeddedStructs_NonStruct(t *testing.T) {
	// Test that non-struct embedded fields are handled gracefully
	type CustomInt int

	type Outer struct {
		CustomInt        // Embedded non-struct type
		Name      string `schema:"name"`
	}

	t.Run("embedded non-struct type", func(t *testing.T) {
		data := map[string]any{
			"name": "Test",
		}

		var result Outer
		u := NewDefaultUnmarshaler()
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Equal(t, "Test", result.Name)
		// CustomInt should remain at zero value since it's not a struct
		assert.Equal(t, CustomInt(0), result.CustomInt)
	})
}

func TestUnmarshaler_Unmarshal_SliceFastPaths(t *testing.T) {
	type ResultIntSlice struct {
		Items []int
	}

	type ResultInterfaceSlice struct {
		Items []interface{}
	}

	type ResultByteSlice struct {
		Data []byte
	}

	type ResultStringSlice struct {
		Names []string
	}

	u := NewDefaultUnmarshaler()

	t.Run("fast path: assignable slice types", func(t *testing.T) {
		// Direct int slice should use fast path
		data := map[string]any{"Items": []int{1, 2, 3, 4, 5}}
		var result ResultIntSlice
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result.Items)
	})

	t.Run("fast path: interface slice", func(t *testing.T) {
		// Slice with interface{} elements
		data := map[string]any{"Items": []any{"string", 42, true, 3.14}}
		var result ResultInterfaceSlice
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Len(t, result.Items, 4)
		assert.Equal(t, "string", result.Items[0])
		assert.Equal(t, 42, result.Items[1])
		assert.Equal(t, true, result.Items[2])
		assert.Equal(t, 3.14, result.Items[3])
	})

	t.Run("fast path: byte slice direct", func(t *testing.T) {
		// Direct byte slice assignment
		data := map[string]any{"Data": []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}}
		var result ResultByteSlice
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Equal(t, []byte("Hello"), result.Data)
	})

	t.Run("fast path: string slice direct", func(t *testing.T) {
		// Direct string slice
		data := map[string]any{"Names": []string{"Alice", "Bob", "Charlie"}}
		var result ResultStringSlice
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Equal(t, []string{"Alice", "Bob", "Charlie"}, result.Names)
	})

	t.Run("conversion path: mixed types in slice", func(t *testing.T) {
		// Mixed types requiring conversion
		data := map[string]any{"Items": []any{1, "2", 3.0, true}}
		var result ResultIntSlice
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3, 1}, result.Items)
	})

	t.Run("empty slice fast path", func(t *testing.T) {
		data := map[string]any{"Items": []int{}}
		var result ResultIntSlice
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Empty(t, result.Items)
	})

	t.Run("large slice performance", func(t *testing.T) {
		// Test with larger slice to ensure fast paths work at scale
		largeSlice := make([]int, 1000)
		for i := range largeSlice {
			largeSlice[i] = i
		}

		data := map[string]any{"Items": largeSlice}
		var result ResultIntSlice
		err := u.Unmarshal(data, &result)

		require.NoError(t, err)
		assert.Len(t, result.Items, 1000)
		assert.Equal(t, 0, result.Items[0])
		assert.Equal(t, 999, result.Items[999])
	})
}
