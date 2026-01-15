package mapstructure

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_convertBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    []byte
		wantErr bool
	}{
		{name: "nil", input: nil, want: nil},
		{name: "bytes", input: []byte{1, 2, 3}, want: []byte{1, 2, 3}},
		{name: "bytes empty", input: []byte{}, want: []byte{}},
		{name: "string", input: "hello", want: []byte("hello")},
		{name: "string empty", input: "", want: []byte("")},
		{name: "any slice", input: []any{1, 2, 3}, want: []byte{1, 2, 3}},
		{name: "any slice empty", input: []any{}, want: []byte{}},
		{name: "any slice with ints", input: []any{int(65), int(66), int(67)}, want: []byte{65, 66, 67}},
		{name: "reader", input: io.NopCloser(strings.NewReader("test")), want: []byte("test")},
		{name: "reader empty", input: io.NopCloser(strings.NewReader("")), want: []byte{}},
		{name: "invalid int", input: 42, wantErr: true},
		{name: "invalid bool", input: true, wantErr: true},
		{name: "any slice invalid element", input: []any{1, "invalid", 3}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertBytes(tt.input)
			if tt.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:forcetypeassert // Test code
			assert.Equal(t, tt.want, result.Interface().([]byte))
		})
	}
}

func TestConverter_convertReadCloser(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		wantContent string
		wantNil     bool
		wantErr     bool
	}{
		{name: "nil", input: nil, wantNil: true},
		{name: "readcloser", input: io.NopCloser(strings.NewReader("hello")), wantContent: "hello"},
		{name: "reader", input: strings.NewReader("world"), wantContent: "world"},
		{name: "bytes", input: []byte("bytes"), wantContent: "bytes"},
		{name: "bytes empty", input: []byte{}, wantContent: ""},
		{name: "string", input: "string", wantContent: "string"},
		{name: "string empty", input: "", wantContent: ""},
		{name: "invalid int", input: 42, wantErr: true},
		{name: "invalid bool", input: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertReadCloser(tt.input)
			if tt.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)

			if tt.wantNil {
				assert.True(t, result.IsNil())

				return
			}

			//nolint:forcetypeassert // Test code
			rc := result.Interface().(io.ReadCloser)
			defer func() { _ = rc.Close() }()

			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, string(content))
		})
	}
}

// errorReader is a test helper that always returns an error on Read.
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func TestConverter_convertBytes_ReadErrors(t *testing.T) {
	t.Run("reader error", func(t *testing.T) {
		errReader := &errorReader{err: io.ErrUnexpectedEOF}
		_, err := convertBytes(errReader)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read")
		assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
	})

	t.Run("readcloser error", func(t *testing.T) {
		errRC := io.NopCloser(&errorReader{err: io.ErrClosedPipe})
		_, err := convertBytes(errRC)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read")
		assert.ErrorIs(t, err, io.ErrClosedPipe)
	})

	t.Run("reader with custom error", func(t *testing.T) {
		customErr := assert.AnError
		errReader := &errorReader{err: customErr}
		_, err := convertBytes(errReader)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read")
	})

	t.Run("reader partial read then error", func(t *testing.T) {
		// Reader that returns EOF immediately
		errReader := &errorReader{err: io.EOF}
		result, err := convertBytes(errReader)

		// io.ReadAll handles EOF as success with 0 bytes
		require.NoError(t, err)
		//nolint:forcetypeassert // Test code
		bytes := result.Interface().([]byte)
		assert.Empty(t, bytes)
	})
}

func TestConverter_convertBytes_EdgeCases(t *testing.T) {
	t.Run("reader then readcloser conversion", func(t *testing.T) {
		// Test that io.Reader is handled before checking io.ReadCloser
		reader := strings.NewReader("test content")
		result, err := convertBytes(reader)

		require.NoError(t, err)
		//nolint:forcetypeassert // Test code
		bytes := result.Interface().([]byte)
		assert.Equal(t, []byte("test content"), bytes)
	})

	t.Run("large reader content", func(t *testing.T) {
		// Test reading larger content
		largeContent := strings.Repeat("x", 10000)
		reader := strings.NewReader(largeContent)
		result, err := convertBytes(reader)

		require.NoError(t, err)
		//nolint:forcetypeassert // Test code
		bytes := result.Interface().([]byte)
		assert.Len(t, bytes, 10000)
		assert.Equal(t, []byte(largeContent), bytes)
	})
}
