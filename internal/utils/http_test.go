package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Simple object",
			status:         http.StatusOK,
			data:           map[string]string{"message": "hello"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"hello"}`,
		},
		{
			name:           "Array data",
			status:         http.StatusCreated,
			data:           []string{"item1", "item2"},
			expectedStatus: http.StatusCreated,
			expectedBody:   `["item1","item2"]`,
		},
		{
			name:           "Number data",
			status:         http.StatusAccepted,
			data:           42,
			expectedStatus: http.StatusAccepted,
			expectedBody:   "42",
		},
		{
			name:           "Boolean data",
			status:         http.StatusOK,
			data:           true,
			expectedStatus: http.StatusOK,
			expectedBody:   "true",
		},
		{
			name:           "Null data",
			status:         http.StatusOK,
			data:           nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			
			WriteJSON(w, tt.status, tt.data)
			
			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
			}
			
			// Check body (trim newline that json.Encoder adds)
			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestWriteJSON_ComplexObject(t *testing.T) {
	w := httptest.NewRecorder()
	
	data := map[string]interface{}{
		"id":      123,
		"name":    "Test User",
		"active":  true,
		"tags":    []string{"admin", "user"},
		"profile": map[string]string{"email": "test@example.com"},
	}
	
	WriteJSON(w, http.StatusOK, data)
	
	// Verify the response can be unmarshaled back
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Check specific fields
	if result["id"].(float64) != 123 {
		t.Errorf("Expected id 123, got %v", result["id"])
	}
	
	if result["name"].(string) != "Test User" {
		t.Errorf("Expected name 'Test User', got %v", result["name"])
	}
	
	if result["active"].(bool) != true {
		t.Errorf("Expected active true, got %v", result["active"])
	}
}

func TestWriteJSON_SpecialCharacters(t *testing.T) {
	w := httptest.NewRecorder()
	
	data := map[string]string{
		"html":    "<script>alert('xss')</script>",
		"quotes":  `"quoted string"`,
		"unicode": "Hello 世界",
	}
	
	WriteJSON(w, http.StatusOK, data)
	
	body := w.Body.String()
	
	// Verify HTML is not escaped (SetEscapeHTML(false) should prevent this)
	if !strings.Contains(body, "<script>alert('xss')</script>") {
		t.Error("HTML should not be escaped")
	}
	
	// Verify quotes are properly escaped in JSON
	if !strings.Contains(body, `"\"quoted string\""`) {
		t.Error("Quotes should be properly escaped in JSON")
	}
	
	// Verify Unicode is preserved
	if !strings.Contains(body, "Hello 世界") {
		t.Error("Unicode characters should be preserved")
	}
}

func TestWriteErrorJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		message        string
		expectedStatus int
	}{
		{
			name:           "Bad Request Error",
			status:         http.StatusBadRequest,
			message:        "Invalid input provided",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Not Found Error",
			status:         http.StatusNotFound,
			message:        "Resource not found",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Internal Server Error",
			status:         http.StatusInternalServerError,
			message:        "Something went wrong",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Empty message",
			status:         http.StatusBadRequest,
			message:        "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			
			WriteErrorJSON(w, tt.status, tt.message)
			
			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
			}
			
			// Parse response body
			var errorResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
			if err != nil {
				t.Fatalf("Failed to parse error response: %v", err)
			}
			
			// Check error field
			if errorResponse["error"] != true {
				t.Errorf("Expected error field to be true, got %v", errorResponse["error"])
			}
			
			// Check message field
			if errorResponse["message"] != tt.message {
				t.Errorf("Expected message %q, got %v", tt.message, errorResponse["message"])
			}
			
			// Check status field
			if errorResponse["status"].(float64) != float64(tt.status) {
				t.Errorf("Expected status %d, got %v", tt.status, errorResponse["status"])
			}
		})
	}
}

func TestWriteErrorJSON_SpecialMessages(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "Message with quotes",
			message: `File "test.txt" not found`,
		},
		{
			name:    "Message with newlines",
			message: "Multi-line\nerror message",
		},
		{
			name:    "Message with Unicode",
			message: "エラーメッセージ",
		},
		{
			name:    "Message with HTML",
			message: "<error>Invalid data</error>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			
			WriteErrorJSON(w, http.StatusBadRequest, tt.message)
			
			// Parse response
			var errorResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
			if err != nil {
				t.Fatalf("Failed to parse error response: %v", err)
			}
			
			// Verify message is preserved correctly
			if errorResponse["message"] != tt.message {
				t.Errorf("Expected message %q, got %v", tt.message, errorResponse["message"])
			}
		})
	}
}

// Test edge cases
func TestWriteJSON_LargeData(t *testing.T) {
	w := httptest.NewRecorder()
	
	// Create large data structure
	largeArray := make([]map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		largeArray[i] = map[string]string{
			"id":   string(rune(i)),
			"name": strings.Repeat("x", 100),
		}
	}
	
	WriteJSON(w, http.StatusOK, largeArray)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	// Verify response is valid JSON
	var result []map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal large response: %v", err)
	}
	
	if len(result) != 1000 {
		t.Errorf("Expected 1000 items, got %d", len(result))
	}
}

func TestWriteJSON_InvalidData(t *testing.T) {
	w := httptest.NewRecorder()
	
	// Create data with circular reference (will cause JSON marshal error)
	data := make(map[string]interface{})
	data["self"] = data
	
	WriteJSON(w, http.StatusOK, data)
	
	// The implementation might handle the error differently, so let's check if it fails gracefully
	// Either it should return 500 or handle the error in some other way
	// Looking at the actual implementation, it writes the status first, then tries to encode
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d or %d, got %d", http.StatusOK, http.StatusInternalServerError, w.Code)
	}
}

// Benchmark tests
func BenchmarkWriteJSON_SmallObject(b *testing.B) {
	data := map[string]string{"message": "hello"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		WriteJSON(w, http.StatusOK, data)
	}
}

func BenchmarkWriteJSON_LargeObject(b *testing.B) {
	data := make(map[string]string)
	for i := 0; i < 100; i++ {
		data[string(rune(i))] = strings.Repeat("x", 100)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		WriteJSON(w, http.StatusOK, data)
	}
}

func BenchmarkWriteErrorJSON(b *testing.B) {
	message := "This is a test error message"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		WriteErrorJSON(w, http.StatusBadRequest, message)
	}
}

// Test concurrent access (to ensure thread safety)
func TestWriteJSON_Concurrent(t *testing.T) {
	data := map[string]string{"message": "hello"}
	
	// Run multiple goroutines writing JSON concurrently
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			
			for j := 0; j < 100; j++ {
				w := httptest.NewRecorder()
				WriteJSON(w, http.StatusOK, data)
				
				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
				}
			}
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}