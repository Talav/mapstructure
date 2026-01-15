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

func TestNewTagCacheBuilder(t *testing.T) {
	type TestStruct struct {
		Name     string `schema:"name"`
		Age      int    `schema:"age"`
		Ignored  string `schema:"-"`
		NoTag    string
		JSONOnly string `json:"json_field"`
	}

	t.Run("schema tag", func(t *testing.T) {
		builder := NewTagCacheBuilder("schema")
		metadata, err := builder(reflect.TypeOf(TestStruct{}))
		require.NoError(t, err)

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
		builder := NewTagCacheBuilder("json")
		metadata, err := builder(reflect.TypeOf(TestStruct{}))
		require.NoError(t, err)

		fieldMap := make(map[string]string)
		for _, f := range metadata.Fields {
			fieldMap[f.StructFieldName] = f.MapKey
		}

		// JSONOnly should use json tag value
		assert.Equal(t, "json_field", fieldMap["JSONOnly"])
		// Others use field name (no json tag)
		assert.Equal(t, "Name", fieldMap["Name"])
	})
}

func TestDefaultCacheBuilder_UsesSchemaTag(t *testing.T) {
	type TestStruct struct {
		Field1 string `schema:"custom_field"`
		Field2 int    `schema:"-"`
		Field3 bool
	}

	metadata, err := DefaultCacheBuilder(reflect.TypeOf(TestStruct{}))
	require.NoError(t, err)

	assert.Len(t, metadata.Fields, 2)

	fieldMap := make(map[string]string)
	for _, f := range metadata.Fields {
		fieldMap[f.StructFieldName] = f.MapKey
	}

	assert.Equal(t, "custom_field", fieldMap["Field1"])
	assert.Equal(t, "Field3", fieldMap["Field3"])
	_, hasField2 := fieldMap["Field2"]
	assert.False(t, hasField2)
}

func TestNewTagCacheBuilder_EmbeddedStruct(t *testing.T) {
	type Inner struct {
		Value string `schema:"inner_value"`
	}

	type Outer struct {
		Inner
		Name string `schema:"name"`
	}

	builder := NewTagCacheBuilder("schema")
	metadata, err := builder(reflect.TypeOf(Outer{}))
	require.NoError(t, err)

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
}

func TestNewTagCacheBuilder_UnexportedFieldsIgnored(t *testing.T) {
	type TestStruct struct {
		Exported   string `schema:"exported"`
		unexported string `schema:"unexported"` //nolint:unused // intentionally testing unexported
	}

	builder := NewTagCacheBuilder("schema")
	metadata, err := builder(reflect.TypeOf(TestStruct{}))
	require.NoError(t, err)

	assert.Len(t, metadata.Fields, 1)
	assert.Equal(t, "Exported", metadata.Fields[0].StructFieldName)
}

func TestNewStructMetadataCache_NilBuilder(t *testing.T) {
	type TestStruct struct {
		Name string `schema:"name"`
		Age  int    `schema:"age"`
	}

	// Create cache with nil builder - should use DefaultCacheBuilder
	cache := NewStructMetadataCache(nil)

	metadata, err := cache.getStructMetadata(reflect.TypeOf(TestStruct{}))
	require.NoError(t, err)
	require.NotNil(t, metadata)

	// Should have parsed schema tags using default builder
	assert.Len(t, metadata.Fields, 2)

	fieldMap := make(map[string]string)
	for _, f := range metadata.Fields {
		fieldMap[f.StructFieldName] = f.MapKey
	}

	assert.Equal(t, "name", fieldMap["Name"])
	assert.Equal(t, "age", fieldMap["Age"])
}

func TestStructMetadataCache_GetStructMetadata_Caching(t *testing.T) {
	type TestStruct struct {
		Field1 string `schema:"field1"`
	}

	cache := NewStructMetadataCache(DefaultCacheBuilder)
	typ := reflect.TypeOf(TestStruct{})

	// First call - should build metadata
	metadata1, err := cache.getStructMetadata(typ)
	require.NoError(t, err)
	require.NotNil(t, metadata1)

	// Second call - should return cached metadata
	metadata2, err := cache.getStructMetadata(typ)
	require.NoError(t, err)
	require.NotNil(t, metadata2)

	// Should be the same instance from cache
	assert.Len(t, metadata1.Fields, 1)
	assert.Len(t, metadata2.Fields, 1)
	assert.Equal(t, "field1", metadata1.Fields[0].MapKey)
	assert.Equal(t, "field1", metadata2.Fields[0].MapKey)
}
