# mapstructure

[![Go Reference](https://pkg.go.dev/badge/github.com/talav/mapstructure.svg)](https://pkg.go.dev/github.com/talav/mapstructure)
[![Go Report Card](https://goreportcard.com/badge/github.com/talav/mapstructure)](https://goreportcard.com/report/github.com/talav/mapstructure)
[![CI](https://github.com/talav/mapstructure/actions/workflows/mapstructure-ci.yml/badge.svg)](https://github.com/talav/mapstructure/actions)
[![codecov](https://codecov.io/gh/Talav/mapstructure/graph/badge.svg?token=ahPYV4ORx0)](https://codecov.io/gh/Talav/mapstructure)

Go library for decoding `map[string]any` values into strongly-typed structs with automatic type conversion and comprehensive struct tag support.

## Features

* ✅ **Automatic type conversion** - string ↔ int, bool, float, and more
* ✅ **Struct tag support** - Flexible field name mapping with `schema`, `json`, or custom tags
* ✅ **Nested structs** - Deep nesting and embedded struct handling
* ✅ **Default values** - Set defaults via `default` tag
* ✅ **Custom converters** - Register converters for custom types
* ✅ **Thread-safe** - Concurrent unmarshaling with cached metadata
* ✅ **Zero allocations** - Efficient slice and struct unmarshaling
* ✅ **Battle-tested** - 90%+ test coverage with comprehensive edge cases
* ✅ **Minimal dependencies** - Only 1 runtime dependency ([tagparser](https://github.com/talav/tagparser))

## Installation

```bash
go get github.com/talav/mapstructure
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/talav/mapstructure"
)

type Person struct {
    Name string `schema:"name"`
    Age  int    `schema:"age"`
}

func main() {
    data := map[string]any{
        "name": "Alice",
        "age":  "30", // string automatically converted to int
    }

    var person Person
    if err := mapstructure.Unmarshal(data, &person); err != nil {
        panic(err)
    }

    fmt.Printf("%+v\n", person) // {Name:Alice Age:30}
}
```

## Usage

### Basic Unmarshaling

The simplest way to use mapstructure is with the convenience function:

```go
type Config struct {
    Host string `schema:"host"`
    Port int    `schema:"port"`
}

data := map[string]any{
    "host": "localhost",
    "port": 8080,
}

var config Config
err := mapstructure.Unmarshal(data, &config)
```

### Struct Tags

By default, the `schema` tag is used for field mapping:

```go
type Config struct {
    ServerHost string `schema:"server_host"`
    ServerPort int    `schema:"port"`
    Debug      bool   `schema:"debug"`
    Ignored    string `schema:"-"` // Skip this field
}
```

| Tag | Behavior |
|-----|----------|
| `schema:"name"` | Use "name" as the map key |
| `schema:"-"` | Skip field entirely |
| No tag | Use Go field name |

### Type Conversion

Built-in converters handle common type conversions automatically:

| Target Type | Accepted Input Types | Example |
|-------------|---------------------|---------|
| `string` | string, bool, int, uint, float, []byte | `42` → `"42"` |
| `bool` | bool, int, uint, float, string | `"true"`, `1` → `true` |
| `int`, `int8`...`int64` | int, uint, float, bool, string | `"42"` → `42` |
| `uint`, `uint8`...`uint64` | int, uint, float, bool, string | `"42"` → `uint(42)` |
| `float32`, `float64` | int, uint, float, bool, string | `"3.14"` → `3.14` |
| `[]byte` | []byte, string, []any, io.Reader | `"Hello"` → `[]byte("Hello")` |
| `io.ReadCloser` | io.ReadCloser, io.Reader, []byte, string | Wraps in `io.NopCloser` |

**Type conversion examples:**

```go
type Example struct {
    Count   int     `schema:"count"`
    Price   float64 `schema:"price"`
    Enabled bool    `schema:"enabled"`
}

// All of these work:
data := map[string]any{
    "count":   "42",    // string → int
    "price":   100,     // int → float64
    "enabled": 1,       // int → bool
}

var ex Example
mapstructure.Unmarshal(data, &ex)
// ex.Count = 42, ex.Price = 100.0, ex.Enabled = true
```

### Default Values

Use the `default` tag to set default values for missing fields:

```go
type Config struct {
    Host    string `schema:"host" default:"localhost"`
    Port    int    `schema:"port" default:"8080"`
    Debug   bool   `schema:"debug" default:"false"`
    Timeout int    `schema:"timeout" default:"30"`
}

data := map[string]any{
    "host": "example.com",
    // port, debug, timeout are missing
}

var config Config
mapstructure.Unmarshal(data, &config)
// config.Host = "example.com"
// config.Port = 8080 (from default)
// config.Debug = false (from default)
// config.Timeout = 30 (from default)
```

### Nested Structs

Nested structs are handled automatically:

```go
type Address struct {
    City    string `schema:"city"`
    Country string `schema:"country"`
}

type Person struct {
    Name    string  `schema:"name"`
    Address Address `schema:"address"`
}

data := map[string]any{
    "name": "Alice",
    "address": map[string]any{
        "city":    "New York",
        "country": "USA",
    },
}

var person Person
mapstructure.Unmarshal(data, &person)
```

### Embedded Structs

Embedded structs support both promoted and named field access:

```go
type Timestamps struct {
    CreatedAt string `schema:"created_at"`
    UpdatedAt string `schema:"updated_at"`
}

type User struct {
    Timestamps        // Embedded - fields promoted to parent
    Name       string `schema:"name"`
}

// Option 1: Promoted fields (flat structure)
data1 := map[string]any{
    "name":       "Alice",
    "created_at": "2024-01-01",
    "updated_at": "2024-01-02",
}

// Option 2: Named embedded access (nested)
data2 := map[string]any{
    "name": "Alice",
    "Timestamps": map[string]any{
        "created_at": "2024-01-01",
        "updated_at": "2024-01-02",
    },
}

// Both work!
var user User
mapstructure.Unmarshal(data1, &user) // Promoted fields
mapstructure.Unmarshal(data2, &user) // Named access
```

### Pointers and Slices

```go
type Config struct {
    Tags    []string `schema:"tags"`
    Count   *int     `schema:"count"`
    Data    []byte   `schema:"data"`
}

data := map[string]any{
    "tags":  []any{"go", "api"},
    "count": 42,
    "data":  []any{72, 101, 108, 108, 111}, // Converts to []byte("Hello")
}

var config Config
mapstructure.Unmarshal(data, &config)
```

**⚠️ Detecting missing fields:**

By default, missing fields get zero values. Use pointers to distinguish missing from zero:

```go
type Config struct {
    Port    *int  `schema:"port"`    // nil if missing
    Enabled *bool `schema:"enabled"` // nil if missing
}

data := map[string]any{
    "port": 8080,
    // enabled is missing
}

var config Config
mapstructure.Unmarshal(data, &config)

if config.Port != nil {
    fmt.Println(*config.Port) // 8080
}

if config.Enabled == nil {
    fmt.Println("enabled not provided") // This prints
}
```

### Custom Tag Names

Use a different tag (e.g., `json`, `yaml`, `db`):

```go
cache := mapstructure.NewStructMetadataCache("json", "default")
converters := mapstructure.NewDefaultConverterRegistry()
unmarshaler := mapstructure.NewUnmarshaler(cache, converters)

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

data := map[string]any{"name": "Alice", "email": "alice@example.com"}
var user User
unmarshaler.Unmarshal(data, &user)
```

**Use custom default value tags:**

```go
// Use "dflt" tag for default values instead of "default"
cache := mapstructure.NewStructMetadataCache("schema", "dflt")
unmarshaler := mapstructure.NewUnmarshaler(cache, mapstructure.NewDefaultConverterRegistry())

type Config struct {
    Host string `schema:"host" dflt:"localhost"`
    Port int    `schema:"port" dflt:"8080"`
}

// Missing values will use defaults from "dflt" tag
data := map[string]any{}
var config Config
unmarshaler.Unmarshal(data, &config)
// Result: Config{Host: "localhost", Port: 8080}
```

**Cleaner API using defaults:**

```go
// Uses "schema" + "default" tags
cache := mapstructure.NewDefaultStructMetadataCache()
unmarshaler := mapstructure.NewUnmarshaler(cache, mapstructure.NewDefaultConverterRegistry())

// Or even simpler:
unmarshaler := mapstructure.NewDefaultUnmarshaler()
```

### Custom Converters

Register converters for custom types:

```go
import (
    "reflect"
    "time"
)

// Define a converter for time.Time
timeConverter := func(value any) (reflect.Value, error) {
    s, ok := value.(string)
    if !ok {
        return reflect.Value{}, fmt.Errorf("expected string for time.Time")
    }
    t, err := time.Parse(time.RFC3339, s)
    if err != nil {
        return reflect.Value{}, err
    }
    return reflect.ValueOf(t), nil
}

// Register the converter
converters := mapstructure.NewDefaultConverterRegistry(map[reflect.Type]mapstructure.Converter{
    reflect.TypeOf(time.Time{}): timeConverter,
})

cache := mapstructure.NewStructMetadataCache("schema", "default")
unmarshaler := mapstructure.NewUnmarshaler(cache, converters)

type Event struct {
    Name      string    `schema:"name"`
    Timestamp time.Time `schema:"timestamp"`
}

data := map[string]any{
    "name":      "meeting",
    "timestamp": "2024-01-15T10:30:00Z",
}

var event Event
unmarshaler.Unmarshal(data, &event)
```

**Custom converter for enums:**

```go
type Status int

const (
    StatusPending Status = iota
    StatusActive
    StatusClosed
)

statusConverter := func(value any) (reflect.Value, error) {
    s, ok := value.(string)
    if !ok {
        return reflect.Value{}, fmt.Errorf("expected string")
    }
    
    switch s {
    case "pending":
        return reflect.ValueOf(StatusPending), nil
    case "active":
        return reflect.ValueOf(StatusActive), nil
    case "closed":
        return reflect.ValueOf(StatusClosed), nil
    default:
        return reflect.Value{}, fmt.Errorf("unknown status: %s", s)
    }
}
```

## Real-World Examples

### API Response Parsing

```go
type APIResponse struct {
    Status  string          `schema:"status"`
    Code    int             `schema:"code"`
    Message string          `schema:"message"`
    Data    json.RawMessage `schema:"data"`
}

type UserData struct {
    ID    int    `schema:"id"`
    Name  string `schema:"name"`
    Email string `schema:"email"`
}

// Parse outer response
var response APIResponse
mapstructure.Unmarshal(apiData, &response)

// Parse nested data if needed
if response.Status == "success" {
    var userData map[string]any
    json.Unmarshal(response.Data, &userData)
    
    var user UserData
    mapstructure.Unmarshal(userData, &user)
}
```

## Error Handling

The library provides structured error types for better error handling:

### Structured Error Types

**ConversionError** - Type conversion failures:

```go
type Config struct {
    Port int `schema:"port"`
}

data := map[string]any{
    "port": "not-a-number",
}

var config Config
err := mapstructure.Unmarshal(data, &config)
if err != nil {
    // Check if it's a conversion error
    var convErr *mapstructure.ConversionError
    if errors.As(err, &convErr) {
        fmt.Printf("Field: %s\n", convErr.FieldPath)      // "port"
        fmt.Printf("Value: %v\n", convErr.Value)          // "not-a-number"
        fmt.Printf("Target: %v\n", convErr.TargetType)    // int
        fmt.Printf("Cause: %v\n", convErr.Cause)          // parsing error
    }
}
```

**ValidationError** - Input validation failures:

```go
// Non-pointer error
err := mapstructure.Unmarshal(data, Config{}) // Wrong!
var valErr *mapstructure.ValidationError
if errors.As(err, &valErr) {
    fmt.Println(valErr.Message) // "result must be a pointer"
}

// Nil pointer error
var config *Config
err := mapstructure.Unmarshal(data, config)
// ValidationError: "result pointer is nil"
```

**Error messages include field paths:**

```go
type Nested struct {
    Inner struct {
        Value int `schema:"value"`
    } `schema:"inner"`
}

data := map[string]any{
    "inner": map[string]any{
        "value": "invalid",
    },
}

var nested Nested
err := mapstructure.Unmarshal(data, &nested)
// Error: inner.value: cannot convert string to int
```

## Performance

The library is optimized for production use:

**Key optimizations:**
- ✅ **Struct metadata caching** - Reflection done once per type
- ✅ **Fast-path slice operations** - Zero-copy for compatible types
- ✅ **Immutable converter registry** - Lock-free concurrent reads


## Thread Safety

**Safe for concurrent use:**
- ✅ `StructMetadataCache` uses `sync.Map` for concurrent access
- ✅ `ConverterRegistry` is immutable after construction
- ✅ `Unmarshaler` is safe for concurrent unmarshaling

```go
// Safe: Shared unmarshaler across goroutines
var unmarshaler = mapstructure.NewDefaultUnmarshaler()

func handler1() {
    var result1 Type1
    unmarshaler.Unmarshal(data1, &result1) // Concurrent safe
}

func handler2() {
    var result2 Type2
    unmarshaler.Unmarshal(data2, &result2) // Concurrent safe
}
```

## Testing


```bash
# Run tests
go test -v

# Run with race detector
go test -race

# Run with coverage
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## API Reference

### Top-Level Functions

```go
// Unmarshal transforms map[string]any into a Go struct
// Uses default settings (schema tag, standard converters)
func Unmarshal(data map[string]any, result any) error
```

### Types

```go
// Unmarshaler handles unmarshaling with custom configuration
type Unmarshaler struct { /* ... */ }

// Converter converts a value to a reflect.Value
type Converter func(value any) (reflect.Value, error)

// ConverterRegistry manages type converters
type ConverterRegistry struct { /* ... */ }

// StructMetadataCache caches struct field metadata
type StructMetadataCache struct { /* ... */ }

// FieldMetadata holds cached struct field information
type FieldMetadata struct {
    StructFieldName string
    MapKey          string
    Index           int
    Type            reflect.Type
    Embedded        bool
    Default         *string
}
```

### Constructors

```go
// NewUnmarshaler creates a new unmarshaler with explicit dependencies
func NewUnmarshaler(cache *StructMetadataCache, converters *ConverterRegistry) *Unmarshaler

// NewDefaultUnmarshaler creates an unmarshaler with default settings
func NewDefaultUnmarshaler() *Unmarshaler

// NewStructMetadataCache creates a metadata cache
// tagName specifies which tag to read for field mapping (e.g., "schema", "json", "yaml")
// defaultTagName specifies which tag to read for default values (e.g., "default")
// Use "-" for tagName to ignore tags and map by field names only
// Empty strings default to "schema" and "default" respectively
func NewStructMetadataCache(tagName, defaultTagName string) *StructMetadataCache

// NewDefaultStructMetadataCache creates a cache with default tag names ("schema", "default")
func NewDefaultStructMetadataCache() *StructMetadataCache

// NewDefaultConverterRegistry creates a registry with standard converters
// additional converter maps can override or extend defaults
func NewDefaultConverterRegistry(additional ...map[reflect.Type]Converter) *ConverterRegistry

// NewConverterRegistry creates a registry with only specified converters
func NewConverterRegistry(converters map[reflect.Type]Converter) *ConverterRegistry
```

### Methods

```go
// Unmarshal transforms map[string]any into result struct
func (u *Unmarshaler) Unmarshal(data map[string]any, result any) error

// GetMetadata retrieves or builds cached struct field metadata
// Safe for concurrent use; useful for pre-warming cache or introspection
func (c *StructMetadataCache) GetMetadata(typ reflect.Type) *StructMetadata

// Find looks up a converter for the given type
func (r *ConverterRegistry) Find(typ reflect.Type) (Converter, bool)
```

## Limitations and Best Practices

### Missing Fields vs Zero Values

⚠️ **Important:** By default, missing fields receive Go zero values:

```go
data := map[string]any{"name": "Alice"}
var person Person
mapstructure.Unmarshal(data, &person)
// person.Age = 0 (zero value, not explicitly set!)
```

**Solutions:**

1. **Use pointers** to detect missing fields:
   ```go
   type Person struct {
       Age *int `schema:"age"` // nil if missing
   }
   ```

2. **Use default tags** for explicit defaults:
   ```go
   type Person struct {
       Age int `schema:"age" default:"18"`
   }
   ```

3. **Validate after unmarshaling** if fields are required

### Type Safety

The library performs **best-effort type conversion**:

```go
// These all work (maybe not what you want):
data := map[string]any{
    "age": "abc", // Will fail with error ✓
    "age": 3.14,  // Converts to 3 (truncates)
    "age": true,  // Converts to 1
}
```


## API Stability

This library follows semantic versioning. The public API is stable for v1.x:

**Stable APIs:**
- `Unmarshal()`
- `NewUnmarshaler()`, `NewDefaultUnmarshaler()`
- `NewStructMetadataCache()`, `NewDefaultStructMetadataCache()`
- `NewDefaultConverterRegistry()`, `NewConverterRegistry()`
- All exported types and methods

### Development Commands

```bash
# Run tests
go test -v ./...

# Run with race detector
go test -race ./...

# Run linters
golangci-lint run

# Run benchmarks
go test -bench=. -benchmem

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

Developed by [Talav](https://github.com/talav).

Tag parsing powered by [tagparser](https://github.com/talav/tagparser).

---

**Questions?** Open an issue or discussion on [GitHub](https://github.com/talav/mapstructure).

**Found a bug?** Please report it with a minimal reproduction case.
