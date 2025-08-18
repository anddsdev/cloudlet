package utils

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteJSONData(t *testing.T) {
	tests := []struct {
		name         string
		data         interface{}
		expectedJSON string
		expectError  bool
	}{
		{
			name:         "Simple string",
			data:         "hello world",
			expectedJSON: `"hello world"`,
			expectError:  false,
		},
		{
			name:         "Number",
			data:         42,
			expectedJSON: "42",
			expectError:  false,
		},
		{
			name:         "Boolean true",
			data:         true,
			expectedJSON: "true",
			expectError:  false,
		},
		{
			name:         "Boolean false",
			data:         false,
			expectedJSON: "false",
			expectError:  false,
		},
		{
			name:         "Null value",
			data:         nil,
			expectedJSON: "null",
			expectError:  false,
		},
		{
			name:         "Simple array",
			data:         []int{1, 2, 3},
			expectedJSON: "[1,2,3]",
			expectError:  false,
		},
		{
			name:         "Simple object",
			data:         map[string]string{"key": "value"},
			expectedJSON: `{"key":"value"}`,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteJSONData(&buf, tt.data)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Remove trailing newline that json.Encoder adds
			result := strings.TrimSpace(buf.String())
			if result != tt.expectedJSON {
				t.Errorf("Expected JSON %q, got %q", tt.expectedJSON, result)
			}
		})
	}
}

func TestWriteJSONData_ComplexStructure(t *testing.T) {
	data := map[string]interface{}{
		"string":  "test",
		"number":  123,
		"boolean": true,
		"null":    nil,
		"array":   []string{"a", "b", "c"},
		"object": map[string]int{
			"nested": 456,
		},
	}

	var buf bytes.Buffer
	err := WriteJSONData(&buf, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the result can be unmarshaled back
	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Check each field
	if result["string"] != "test" {
		t.Errorf("Expected string 'test', got %v", result["string"])
	}

	if result["number"].(float64) != 123 {
		t.Errorf("Expected number 123, got %v", result["number"])
	}

	if result["boolean"] != true {
		t.Errorf("Expected boolean true, got %v", result["boolean"])
	}

	if result["null"] != nil {
		t.Errorf("Expected null value, got %v", result["null"])
	}

	array := result["array"].([]interface{})
	if len(array) != 3 || array[0] != "a" || array[1] != "b" || array[2] != "c" {
		t.Errorf("Expected array [a,b,c], got %v", array)
	}

	object := result["object"].(map[string]interface{})
	if object["nested"].(float64) != 456 {
		t.Errorf("Expected nested value 456, got %v", object["nested"])
	}
}

func TestWriteJSONData_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		data         interface{}
		expectedJSON string
	}{
		{
			name:         "HTML characters",
			data:         "<script>alert('xss')</script>",
			expectedJSON: `"<script>alert('xss')</script>"`,
		},
		{
			name:         "Quotes",
			data:         `"quoted string"`,
			expectedJSON: `"\"quoted string\""`,
		},
		{
			name:         "Backslashes",
			data:         `C:\Windows\System32`,
			expectedJSON: `"C:\\Windows\\System32"`,
		},
		{
			name:         "Unicode characters",
			data:         "Hello 世界 🌍",
			expectedJSON: `"Hello 世界 🌍"`,
		},
		{
			name:         "Control characters",
			data:         "line1\nline2\ttab",
			expectedJSON: `"line1\nline2\ttab"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteJSONData(&buf, tt.data)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			result := strings.TrimSpace(buf.String())
			if result != tt.expectedJSON {
				t.Errorf("Expected JSON %q, got %q", tt.expectedJSON, result)
			}
		})
	}
}

func TestWriteJSONData_HTMLEscaping(t *testing.T) {
	// Test that HTML escaping is disabled (SetEscapeHTML(false))
	data := map[string]string{
		"html": "<div>content</div>",
		"url":  "https://example.com/path?param=value&other=123",
	}

	var buf bytes.Buffer
	err := WriteJSONData(&buf, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	result := buf.String()

	// These characters should NOT be escaped
	if !strings.Contains(result, "<div>content</div>") {
		t.Error("HTML tags should not be escaped")
	}

	if !strings.Contains(result, "https://example.com/path?param=value&other=123") {
		t.Error("URL should not be escaped")
	}

	// But quotes should still be escaped for valid JSON
	if !strings.Contains(result, `"<div>content</div>"`) {
		t.Error("String should still be properly quoted in JSON")
	}
}

func TestWriteJSONData_EmptyValues(t *testing.T) {
	tests := []struct {
		name         string
		data         interface{}
		expectedJSON string
	}{
		{
			name:         "Empty string",
			data:         "",
			expectedJSON: `""`,
		},
		{
			name:         "Empty array",
			data:         []string{},
			expectedJSON: "[]",
		},
		{
			name:         "Empty object",
			data:         map[string]string{},
			expectedJSON: "{}",
		},
		{
			name:         "Zero number",
			data:         0,
			expectedJSON: "0",
		},
		{
			name:         "Zero float",
			data:         0.0,
			expectedJSON: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteJSONData(&buf, tt.data)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			result := strings.TrimSpace(buf.String())
			if result != tt.expectedJSON {
				t.Errorf("Expected JSON %q, got %q", tt.expectedJSON, result)
			}
		})
	}
}

func TestWriteJSONData_ErrorCases(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "Circular reference",
			data: func() map[string]interface{} {
				m := make(map[string]interface{})
				m["self"] = m
				return m
			}(),
		},
		{
			name: "Channel (non-serializable)",
			data: make(chan int),
		},
		{
			name: "Function (non-serializable)",
			data: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteJSONData(&buf, tt.data)
			if err == nil {
				t.Error("Expected error for non-serializable data")
			}
		})
	}
}

func TestWriteJSONData_LargeData(t *testing.T) {
	// Test with large data structure
	largeArray := make([]map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		largeArray[i] = map[string]string{
			"id":          string(rune(i)),
			"name":        strings.Repeat("x", 100),
			"description": strings.Repeat("y", 200),
		}
	}

	var buf bytes.Buffer
	err := WriteJSONData(&buf, largeArray)
	if err != nil {
		t.Fatalf("Unexpected error with large data: %v", err)
	}

	// Verify the result is valid JSON
	var result []map[string]string
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal large data: %v", err)
	}

	if len(result) != 1000 {
		t.Errorf("Expected 1000 items, got %d", len(result))
	}

	// Check first and last items
	if len(result[0]["name"]) != 100 {
		t.Errorf("Expected name length 100, got %d", len(result[0]["name"]))
	}

	if len(result[999]["description"]) != 200 {
		t.Errorf("Expected description length 200, got %d", len(result[999]["description"]))
	}
}

// Test different numeric types
func TestWriteJSONData_NumericTypes(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{name: "int", data: int(42)},
		{name: "int8", data: int8(42)},
		{name: "int16", data: int16(42)},
		{name: "int32", data: int32(42)},
		{name: "int64", data: int64(42)},
		{name: "uint", data: uint(42)},
		{name: "uint8", data: uint8(42)},
		{name: "uint16", data: uint16(42)},
		{name: "uint32", data: uint32(42)},
		{name: "uint64", data: uint64(42)},
		{name: "float32", data: float32(42.5)},
		{name: "float64", data: float64(42.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteJSONData(&buf, tt.data)
			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", tt.name, err)
			}

			// Verify result is valid JSON
			var result interface{}
			err = json.Unmarshal(buf.Bytes(), &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal %s: %v", tt.name, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkWriteJSONData_SmallObject(b *testing.B) {
	data := map[string]string{"message": "hello"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		WriteJSONData(&buf, data)
	}
}

func BenchmarkWriteJSONData_LargeObject(b *testing.B) {
	data := make(map[string]string)
	for i := 0; i < 100; i++ {
		data[string(rune(i))] = strings.Repeat("x", 100)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		WriteJSONData(&buf, data)
	}
}

func BenchmarkWriteJSONData_Array(b *testing.B) {
	data := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = i
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		WriteJSONData(&buf, data)
	}
}

// Test concurrent access
func TestWriteJSONData_Concurrent(b *testing.T) {
	data := map[string]string{"message": "hello"}
	
	// Run multiple goroutines concurrently
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			
			for j := 0; j < 100; j++ {
				var buf bytes.Buffer
				err := WriteJSONData(&buf, data)
				if err != nil {
					b.Errorf("Unexpected error: %v", err)
				}
			}
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}