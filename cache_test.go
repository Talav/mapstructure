package mapstructure

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFieldTag(t *testing.T) {
	tests := []struct {
		name      string
		tagValue  string
		fieldName string
		wantKey   string
		wantSkip  bool
	}{
		{
			name:      "empty tag uses field name",
			tagValue:  "",
			fieldName: "MyField",
			wantKey:   "MyField",
			wantSkip:  false,
		},
		{
			name:      "simple name",
			tagValue:  "custom_name",
			fieldName: "MyField",
			wantKey:   "custom_name",
			wantSkip:  false,
		},
		{
			name:      "dash skips field",
			tagValue:  "-",
			fieldName: "MyField",
			wantKey:   "",
			wantSkip:  true,
		},
		{
			name:      "name with options",
			tagValue:  "custom_name,omitempty",
			fieldName: "MyField",
			wantKey:   "custom_name",
			wantSkip:  false,
		},
		{
			name:      "dash with options still skips",
			tagValue:  "-,omitempty",
			fieldName: "MyField",
			wantKey:   "",
			wantSkip:  true,
		},
		{
			name:      "name with key-value option",
			tagValue:  "custom_name,format:date",
			fieldName: "MyField",
			wantKey:   "custom_name",
			wantSkip:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotSkip := parseFieldTag(tt.tagValue, tt.fieldName)
			assert.Equal(t, tt.wantKey, gotKey)
			assert.Equal(t, tt.wantSkip, gotSkip)
		})
	}
}

func TestStructMetadataCache_TagNames(t *testing.T) {
	type TestStruct struct {
		Name     string `schema:"name"`
		Age      int    `schema:"age"`
		Ignored  string `schema:"-"`
		NoTag    string
		JSONOnly string `json:"json_field"`
	}

	t.Run("schema tag", func(t *testing.T) {
		cache := NewStructMetadataCache("schema", "")
		metadata := cache.GetMetadata(reflect.TypeOf(TestStruct{}))

		// Should have 4 fields (Name, Age, NoTag, JSONOnly - Ignored is skipped)
		assert.Len(t, metadata.Fields, 4)

		fieldMap := make(map[string]string)
		for _, f := range metadata.Fields {
			fieldMap[f.StructFieldName] = f.MapKey
		}

		assert.Equal(t, "name", fieldMap["Name"])
		assert.Equal(t, "age", fieldMap["Age"])
		assert.Equal(t, "NoTag", fieldMap["NoTag"])
		assert.Equal(t, "JSONOnly", fieldMap["JSONOnly"]) // No schema tag, uses field name
		_, hasIgnored := fieldMap["Ignored"]
		assert.False(t, hasIgnored)
	})

	t.Run("json tag", func(t *testing.T) {
		cache := NewStructMetadataCache("json", "")
		metadata := cache.GetMetadata(reflect.TypeOf(TestStruct{}))

		fieldMap := make(map[string]string)
		for _, f := range metadata.Fields {
			fieldMap[f.StructFieldName] = f.MapKey
		}

		// JSONOnly should use json tag value
		assert.Equal(t, "json_field", fieldMap["JSONOnly"])
		// Others use field name (no json tag)
		assert.Equal(t, "Name", fieldMap["Name"])
	})

	t.Run("empty string defaults to schema", func(t *testing.T) {
		cache := NewStructMetadataCache("", "")
		metadata := cache.GetMetadata(reflect.TypeOf(TestStruct{}))

		// Should parse as schema tag (default)
		assert.Len(t, metadata.Fields, 4)

		fieldMap := make(map[string]string)
		for _, f := range metadata.Fields {
			fieldMap[f.StructFieldName] = f.MapKey
		}

		assert.Equal(t, "name", fieldMap["Name"])
		assert.Equal(t, "age", fieldMap["Age"])
		_, hasIgnored := fieldMap["Ignored"]
		assert.False(t, hasIgnored, "schema:'-' field should be ignored")
	})

	t.Run("field names only with dash", func(t *testing.T) {
		// "-" means ignore all tags and use field names directly
		cache := NewStructMetadataCache("-", "")
		metadata := cache.GetMetadata(reflect.TypeOf(TestStruct{}))

		// Should have all exported fields (including Ignored)
		assert.Len(t, metadata.Fields, 5)

		fieldMap := make(map[string]string)
		for _, f := range metadata.Fields {
			fieldMap[f.StructFieldName] = f.MapKey
		}

		// All fields should use their struct field names, ignoring tags
		assert.Equal(t, "Name", fieldMap["Name"])
		assert.Equal(t, "Age", fieldMap["Age"])
		assert.Equal(t, "Ignored", fieldMap["Ignored"]) // Not skipped when using "-"
		assert.Equal(t, "NoTag", fieldMap["NoTag"])
		assert.Equal(t, "JSONOnly", fieldMap["JSONOnly"])
	})
}

func TestStructMetadataCache_SpecialFieldTypes(t *testing.T) {
	t.Run("embedded struct", func(t *testing.T) {
		type Inner struct {
			Value string `schema:"inner_value"`
		}

		type Outer struct {
			Inner
			Name string `schema:"name"`
		}

		cache := NewStructMetadataCache("schema", "")
		metadata := cache.GetMetadata(reflect.TypeOf(Outer{}))

		assert.Len(t, metadata.Fields, 2)

		var embeddedField *FieldMetadata
		for i := range metadata.Fields {
			if metadata.Fields[i].StructFieldName == "Inner" {
				embeddedField = &metadata.Fields[i]

				break
			}
		}

		require.NotNil(t, embeddedField)
		assert.True(t, embeddedField.Embedded)
	})

	t.Run("unexported fields ignored", func(t *testing.T) {
		type TestStruct struct {
			Exported   string `schema:"exported"`
			unexported string `schema:"unexported"` //nolint:unused // intentionally testing unexported
		}

		cache := NewStructMetadataCache("schema", "")
		metadata := cache.GetMetadata(reflect.TypeOf(TestStruct{}))

		assert.Len(t, metadata.Fields, 1)
		assert.Equal(t, "Exported", metadata.Fields[0].StructFieldName)
	})
}

func TestStructMetadataCache_Caching(t *testing.T) {
	type TestStruct struct {
		Field1 string `schema:"field1"`
	}

	cache := NewStructMetadataCache("schema", "")
	typ := reflect.TypeOf(TestStruct{})

	// First call - should build metadata
	metadata1 := cache.GetMetadata(typ)
	require.NotNil(t, metadata1)

	// Second call - should return cached metadata
	metadata2 := cache.GetMetadata(typ)
	require.NotNil(t, metadata2)

	// Should be the same instance from cache
	assert.Len(t, metadata1.Fields, 1)
	assert.Len(t, metadata2.Fields, 1)
	assert.Equal(t, "field1", metadata1.Fields[0].MapKey)
	assert.Equal(t, "field1", metadata2.Fields[0].MapKey)
}

func TestNewDefaultStructMetadataCache(t *testing.T) {
	type TestStruct struct {
		Name string `schema:"name" default:"John"`
		Age  int    `schema:"age" default:"30"`
	}

	// Create cache with defaults
	cache := NewDefaultStructMetadataCache()
	metadata := cache.GetMetadata(reflect.TypeOf(TestStruct{}))

	// Should use "schema" tag for field mapping
	fieldMap := make(map[string]string)
	for _, f := range metadata.Fields {
		fieldMap[f.StructFieldName] = f.MapKey
	}
	assert.Equal(t, "name", fieldMap["Name"])
	assert.Equal(t, "age", fieldMap["Age"])

	// Should use "default" tag for default values
	defaultMap := make(map[string]*string)
	for _, f := range metadata.Fields {
		defaultMap[f.StructFieldName] = f.Default
	}
	assert.NotNil(t, defaultMap["Name"])
	assert.Equal(t, "John", *defaultMap["Name"])
	assert.NotNil(t, defaultMap["Age"])
	assert.Equal(t, "30", *defaultMap["Age"])
}

func TestStructMetadataCache_CustomDefaultTag(t *testing.T) {
	type TestStruct struct {
		Name    string `schema:"name" default:"John Doe"`
		Age     int    `schema:"age" dflt:"25"`
		City    string `schema:"city" default:"NYC"`
		Country string `schema:"country" dflt:"USA"`
	}

	t.Run("standard default tag", func(t *testing.T) {
		cache := NewStructMetadataCache("schema", "default")
		metadata := cache.GetMetadata(reflect.TypeOf(TestStruct{}))

		// Check which fields have default values
		defaultMap := make(map[string]*string)
		for _, f := range metadata.Fields {
			defaultMap[f.StructFieldName] = f.Default
		}

		assert.NotNil(t, defaultMap["Name"])
		assert.Equal(t, "John Doe", *defaultMap["Name"])

		assert.Nil(t, defaultMap["Age"]) // Has dflt tag, not default

		assert.NotNil(t, defaultMap["City"])
		assert.Equal(t, "NYC", *defaultMap["City"])

		assert.Nil(t, defaultMap["Country"]) // Has dflt tag, not default
	})

	t.Run("custom default tag", func(t *testing.T) {
		cache := NewStructMetadataCache("schema", "dflt")
		metadata := cache.GetMetadata(reflect.TypeOf(TestStruct{}))

		// Check which fields have default values
		defaultMap := make(map[string]*string)
		for _, f := range metadata.Fields {
			defaultMap[f.StructFieldName] = f.Default
		}

		assert.Nil(t, defaultMap["Name"]) // Has default tag, not dflt

		assert.NotNil(t, defaultMap["Age"])
		assert.Equal(t, "25", *defaultMap["Age"])

		assert.Nil(t, defaultMap["City"]) // Has default tag, not dflt

		assert.NotNil(t, defaultMap["Country"])
		assert.Equal(t, "USA", *defaultMap["Country"])
	})

	t.Run("empty default tag defaults to 'default'", func(t *testing.T) {
		cache := NewStructMetadataCache("schema", "")
		metadata := cache.GetMetadata(reflect.TypeOf(TestStruct{}))

		// Check which fields have default values
		defaultMap := make(map[string]*string)
		for _, f := range metadata.Fields {
			defaultMap[f.StructFieldName] = f.Default
		}

		// Should use "default" tag
		assert.NotNil(t, defaultMap["Name"])
		assert.Equal(t, "John Doe", *defaultMap["Name"])
		assert.NotNil(t, defaultMap["City"])
		assert.Equal(t, "NYC", *defaultMap["City"])
	})
}
