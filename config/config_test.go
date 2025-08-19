package config

import (
	"os"
	"testing"
)

func TestNewConfigFromEnv(t *testing.T) {
	// Set up test environment variables
	testEnvVars := map[string]string{
		"PORT":                        "9090",
		"MAX_MEMORY":                  "64000000",
		"MAX_FILE_SIZE":              "200000000",
		"STORAGE_PATH":               "/test/storage",
		"READ_TIMEOUT":               "45",
		"WRITE_TIMEOUT":              "45",
		"IDLE_TIMEOUT":               "90",
		"MAX_HEADER_BYTES":           "2048",
		"READ_HEADER_TIMEOUT":        "15",
		"MAX_FILES_PER_REQUEST":      "25",
		"MAX_TOTAL_SIZE_PER_REQUEST": "1048576000",
		"ALLOW_PARTIAL_SUCCESS":      "false",
		"ENABLE_BATCH_PROCESSING":    "false",
		"BATCH_SIZE":                 "5",
		"MAX_CONCURRENT_UPLOADS":     "5",
		"STREAMING_THRESHOLD":        "20971520",
		"VALIDATE_BEFORE_UPLOAD":     "false",
		"ENABLE_PROGRESS_TRACKING":   "true",
		"CLEANUP_ON_FAILURE":         "true",
		"RATE_LIMIT_PER_MINUTE":      "200",
		"DB_DSN":                     "/test/db.sqlite",
		"DB_MAX_CONN":                "20",
	}

	// Set environment variables
	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}

	// Clean up after test
	defer func() {
		for key := range testEnvVars {
			os.Unsetenv(key)
		}
	}()

	// Test loading from environment
	config, err := NewConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config from environment: %v", err)
	}

	// Verify server configuration
	if config.Server.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", config.Server.Port)
	}
	if config.Server.MaxMemory != 64000000 {
		t.Errorf("Expected MaxMemory 64000000, got %d", config.Server.MaxMemory)
	}
	if config.Server.MaxFileSize != 200000000 {
		t.Errorf("Expected MaxFileSize 200000000, got %d", config.Server.MaxFileSize)
	}
	if config.Server.Storage.Path != "/test/storage" {
		t.Errorf("Expected storage path /test/storage, got %s", config.Server.Storage.Path)
	}

	// Verify timeout configuration
	if config.Server.Timeout.ReadTimeout != 45 {
		t.Errorf("Expected ReadTimeout 45, got %d", config.Server.Timeout.ReadTimeout)
	}
	if config.Server.Timeout.WriteTimeout != 45 {
		t.Errorf("Expected WriteTimeout 45, got %d", config.Server.Timeout.WriteTimeout)
	}

	// Verify upload configuration
	if config.Server.Upload.MaxFilesPerRequest != 25 {
		t.Errorf("Expected MaxFilesPerRequest 25, got %d", config.Server.Upload.MaxFilesPerRequest)
	}
	if config.Server.Upload.AllowPartialSuccess != false {
		t.Errorf("Expected AllowPartialSuccess false, got %t", config.Server.Upload.AllowPartialSuccess)
	}
	if config.Server.Upload.EnableProgressTracking != true {
		t.Errorf("Expected EnableProgressTracking true, got %t", config.Server.Upload.EnableProgressTracking)
	}

	// Verify database configuration
	if config.Database.DSN != "/test/db.sqlite" {
		t.Errorf("Expected database DSN /test/db.sqlite, got %s", config.Database.DSN)
	}
	if config.Database.MaxConn != 20 {
		t.Errorf("Expected database MaxConn 20, got %d", config.Database.MaxConn)
	}
}

func TestGetEnvString(t *testing.T) {
	// Test with set environment variable
	os.Setenv("TEST_STRING", "test_value")
	defer os.Unsetenv("TEST_STRING")

	result := getEnvString("TEST_STRING", "default")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}

	// Test with unset environment variable
	result = getEnvString("UNSET_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}

func TestGetEnvInt(t *testing.T) {
	// Test with valid integer
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := getEnvInt("TEST_INT", 10)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	// Test with invalid integer
	os.Setenv("TEST_INVALID_INT", "not_a_number")
	defer os.Unsetenv("TEST_INVALID_INT")

	result = getEnvInt("TEST_INVALID_INT", 10)
	if result != 10 {
		t.Errorf("Expected default value 10, got %d", result)
	}

	// Test with unset variable
	result = getEnvInt("UNSET_INT", 15)
	if result != 15 {
		t.Errorf("Expected default value 15, got %d", result)
	}
}

func TestGetEnvBool(t *testing.T) {
	testCases := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"1", true},
		{"yes", true},
		{"YES", true},
		{"on", true},
		{"ON", true},
		{"enabled", true},
		{"ENABLED", true},
		{"false", false},
		{"FALSE", false},
		{"0", false},
		{"no", false},
		{"NO", false},
		{"off", false},
		{"OFF", false},
		{"disabled", false},
		{"DISABLED", false},
		{"invalid", false}, // Should return default for invalid values
	}

	for _, tc := range testCases {
		os.Setenv("TEST_BOOL", tc.value)
		result := getEnvBool("TEST_BOOL", false)
		if result != tc.expected {
			t.Errorf("For value '%s', expected %t, got %t", tc.value, tc.expected, result)
		}
	}

	// Clean up
	os.Unsetenv("TEST_BOOL")

	// Test with unset variable and default true
	result := getEnvBool("UNSET_BOOL", true)
	if result != true {
		t.Errorf("Expected default value true, got %t", result)
	}
}

func TestHasAnyEnvVars(t *testing.T) {
	// Test with no environment variables set
	result := hasAnyEnvVars()
	// This might be true if any env vars are already set, so we'll test by setting one

	// Set one environment variable
	os.Setenv("PORT", "8080")
	defer os.Unsetenv("PORT")

	result = hasAnyEnvVars()
	if !result {
		t.Error("Expected hasAnyEnvVars to return true when PORT is set")
	}
}

func TestNewConfigFallback(t *testing.T) {
	// Test fallback to defaults when no YAML file exists and no env vars are set
	config, err := NewConfig("nonexistent.yaml")
	if err != nil {
		t.Fatalf("Failed to load config with fallback: %v", err)
	}

	// Should have default values
	if config.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", config.Server.Port)
	}
	if config.Server.MaxMemory != 32000000 {
		t.Errorf("Expected default MaxMemory 32000000, got %d", config.Server.MaxMemory)
	}
}

func TestConfigPriority(t *testing.T) {
	// Set an environment variable
	os.Setenv("PORT", "9999")
	defer os.Unsetenv("PORT")

	// Load config (should prioritize env var over YAML)
	config, err := NewConfig("../config/config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Should use environment variable value, not YAML value
	if config.Server.Port != "9999" {
		t.Errorf("Expected environment variable port 9999, got %s", config.Server.Port)
	}
}