package mapstructure

import (
	"reflect"
	"sync"

	"github.com/talav/tagparser"
)

const (
	// DefaultTagName is the default struct tag name for field mapping.
	DefaultTagName = "schema"

	// DefaultValueTagName is the default struct tag name for default values.
	DefaultValueTagName = "default"
)

// StructMetadataCache provides caching for struct field metadata.
type StructMetadataCache struct {
	cache          sync.Map
	tagName        string
	defaultTagName string
}

// NewStructMetadataCache creates a new struct metadata cache.
// tagName specifies which tag to read for field mapping (e.g., "schema", "json", "yaml").
// defaultTagName specifies which tag to read for default values (e.g., "default").
// Use "-" for tagName to ignore all tags and map fields by their Go struct field names.
// Empty strings default to "schema" and "default" respectively.
func NewStructMetadataCache(tagName, defaultTagName string) *StructMetadataCache {
	if tagName == "" {
		tagName = DefaultTagName
	}
	if defaultTagName == "" {
		defaultTagName = DefaultValueTagName
	}

	return &StructMetadataCache{
		tagName:        tagName,
		defaultTagName: defaultTagName,
	}
}

// NewDefaultStructMetadataCache creates a struct metadata cache with default tag names.
// Uses "schema" for field mapping and "default" for default values.
// This is equivalent to NewStructMetadataCache(DefaultTagName, DefaultValueTagName).
func NewDefaultStructMetadataCache() *StructMetadataCache {
	return NewStructMetadataCache(DefaultTagName, DefaultValueTagName)
}

// GetMetadata retrieves or builds cached struct field metadata for the given type.
// This method is safe for concurrent use and will cache the result for subsequent calls.
//
// This is useful for:
//   - Pre-warming the cache before hot paths
//   - Introspecting struct metadata for tooling
//   - Testing cache behavior
func (c *StructMetadataCache) GetMetadata(typ reflect.Type) *StructMetadata {
	// Check cache first
	if cached, ok := c.cache.Load(typ); ok {
		if metadata, ok := cached.(*StructMetadata); ok {
			return metadata
		}
	}

	// Build metadata
	metadata := c.buildMetadata(typ)

	// Store in cache (or get existing if another goroutine stored it first)
	actual, _ := c.cache.LoadOrStore(typ, metadata)
	metadata, _ = actual.(*StructMetadata)

	return metadata
}

// buildMetadata builds struct metadata by parsing struct tags.
func (c *StructMetadataCache) buildMetadata(typ reflect.Type) *StructMetadata {
	fields := make([]FieldMetadata, 0, typ.NumField())

	for i := range typ.NumField() {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}

		// If tagName is "-", use field name directly without reading tags
		var mapKey string
		var skip bool
		if c.tagName == "-" {
			mapKey = f.Name
		} else {
			mapKey, skip = parseFieldTag(f.Tag.Get(c.tagName), f.Name)
			if skip {
				continue
			}
		}

		// Store raw default pointer - conversion happens at unmarshal time
		var defaultPtr *string
		if v, ok := f.Tag.Lookup(c.defaultTagName); ok {
			defaultPtr = &v
		}

		fields = append(fields, FieldMetadata{
			StructFieldName: f.Name,
			MapKey:          mapKey,
			Index:           i,
			Type:            f.Type,
			Embedded:        f.Anonymous,
			Default:         defaultPtr,
		})
	}

	return &StructMetadata{Fields: fields}
}

// parseFieldTag extracts the map key from a tag value.
// Returns (mapKey, skip). If skip is true, the field should be ignored.
func parseFieldTag(tagValue, fieldName string) (string, bool) {
	if tagValue == "" {
		return fieldName, false
	}

	if tagValue == "-" {
		return "", true
	}

	tag, err := tagparser.ParseWithName(tagValue)
	if err != nil || tag.Name == "" {
		return fieldName, false
	}

	if tag.Name == "-" {
		return "", true
	}

	return tag.Name, false
}
