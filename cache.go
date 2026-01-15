package mapstructure

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/talav/tagparser"
)

// CacheBuilderFunc builds struct metadata for caching.
type CacheBuilderFunc func(typ reflect.Type) (*StructMetadata, error)

// DefaultCacheBuilder builds struct metadata using "schema" tags.
var DefaultCacheBuilder = NewTagCacheBuilder("schema")

// NewTagCacheBuilder creates a cache builder that parses struct tags.
// tagName specifies which tag to read (e.g., "schema", "json", "mapstructure").
func NewTagCacheBuilder(tagName string) CacheBuilderFunc {
	return func(typ reflect.Type) (*StructMetadata, error) {
		fields := make([]FieldMetadata, 0, typ.NumField())
		for i := range typ.NumField() {
			f := typ.Field(i)
			if !f.IsExported() {
				continue
			}

			mapKey, skip := parseFieldTag(f.Tag.Get(tagName), f.Name)
			if skip {
				continue
			}

			// Store raw default pointer - conversion happens at unmarshal time
			// using the unmarshaler's converter registry
			var defaultPtr *string
			if v, ok := f.Tag.Lookup("default"); ok {
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

		return &StructMetadata{Fields: fields}, nil
	}
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

// StructMetadataCache provides caching for struct field metadata.
type StructMetadataCache struct {
	cache   sync.Map
	builder CacheBuilderFunc
}

// NewStructMetadataCache creates a new struct metadata cache.
// If builder is nil, DefaultCacheBuilder is used.
func NewStructMetadataCache(builder CacheBuilderFunc) *StructMetadataCache {
	if builder == nil {
		builder = DefaultCacheBuilder
	}

	return &StructMetadataCache{
		builder: builder,
	}
}

// getStructMetadata retrieves or builds cached struct field metadata.
func (c *StructMetadataCache) getStructMetadata(typ reflect.Type) (*StructMetadata, error) {
	// Check cache first
	if cached, ok := c.cache.Load(typ); ok {
		if metadata, ok := cached.(*StructMetadata); ok {
			return metadata, nil
		}
	}

	// Build metadata
	metadata, err := c.builder(typ)
	if err != nil {
		return nil, fmt.Errorf("failed to build struct metadata: %w", err)
	}

	// Store in cache (or get existing if another goroutine stored it first)
	actual, _ := c.cache.LoadOrStore(typ, metadata)
	metadata, _ = actual.(*StructMetadata)

	return metadata, nil
}
