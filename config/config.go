package config

import (
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port        string `yaml:"port"`
		MaxMemory   int    `yaml:"max_memory"`
		MaxFileSize int64  `yaml:"max_file_size"`
		Storage     struct {
			Path string `yaml:"path"`
		} `yaml:"storage"`
		Timeout struct {
			ReadTimeout       int `yaml:"read_timeout"`
			WriteTimeout      int `yaml:"write_timeout"`
			IdleTimeout       int `yaml:"idle_timeout"`
			MaxHeaderBytes    int `yaml:"max_header_bytes"`
			ReadHeaderTimeout int `yaml:"read_header_timeout"`
		} `yaml:"timeout"`
		Upload struct {
			MaxFilesPerRequest     int   `yaml:"max_files_per_request"`
			MaxTotalSizePerRequest int64 `yaml:"max_total_size_per_request"`
			AllowPartialSuccess    bool  `yaml:"allow_partial_success"`
			EnableBatchProcessing  bool  `yaml:"enable_batch_processing"`
			BatchSize              int   `yaml:"batch_size"`
			MaxConcurrentUploads   int   `yaml:"max_concurrent_uploads"`
			StreamingThreshold     int64 `yaml:"streaming_threshold"`
			ValidateBeforeUpload   bool  `yaml:"validate_before_upload"`
			EnableProgressTracking bool  `yaml:"enable_progress_tracking"`
			CleanupOnFailure       bool  `yaml:"cleanup_on_failure"`
			RateLimitPerMinute     int   `yaml:"rate_limit_per_minute"`
		} `yaml:"upload"`
	} `yaml:"server"`

	Database struct {
		DSN     string `yaml:"dsn"`
		MaxConn int    `yaml:"max_conn"`
	} `yaml:"database"`
}

// NewConfig creates a new configuration instance from environment variables
// Falls back to YAML file if environment variables are not set
func NewConfig(fallbackPath string) (*Config, error) {
	config := &Config{}

	// Check if we have any environment variables set
	hasEnvVars := hasAnyEnvVars()

	if hasEnvVars {
		// Load from environment variables (primary source)
		if err := loadFromEnv(config); err != nil {
			return nil, err
		}
	} else {
		// Fallback to YAML if no environment variables are set
		if err := loadFromYAML(config, fallbackPath); err != nil {
			// If YAML also fails, load default values
			loadFromEnv(config) // This will use default values
		}
	}

	return config, nil
}

// NewConfigFromEnv creates a new configuration instance exclusively from environment variables
func NewConfigFromEnv() (*Config, error) {
	config := &Config{}
	return config, loadFromEnv(config)
}

// loadFromEnv populates the config struct from environment variables
func loadFromEnv(config *Config) error {
	// Server configuration
	config.Server.Port = getEnvString("PORT", "8080")
	config.Server.MaxMemory = getEnvInt("MAX_MEMORY", 32000000)
	config.Server.MaxFileSize = getEnvInt64("MAX_FILE_SIZE", 100000000)
	config.Server.Storage.Path = getEnvString("STORAGE_PATH", "./data/storage")

	// Timeout configuration
	config.Server.Timeout.ReadTimeout = getEnvInt("READ_TIMEOUT", 30)
	config.Server.Timeout.WriteTimeout = getEnvInt("WRITE_TIMEOUT", 30)
	config.Server.Timeout.IdleTimeout = getEnvInt("IDLE_TIMEOUT", 60)
	config.Server.Timeout.MaxHeaderBytes = getEnvInt("MAX_HEADER_BYTES", 1024)
	config.Server.Timeout.ReadHeaderTimeout = getEnvInt("READ_HEADER_TIMEOUT", 10)

	// Upload configuration
	config.Server.Upload.MaxFilesPerRequest = getEnvInt("MAX_FILES_PER_REQUEST", 50)
	config.Server.Upload.MaxTotalSizePerRequest = getEnvInt64("MAX_TOTAL_SIZE_PER_REQUEST", 524288000)
	config.Server.Upload.AllowPartialSuccess = getEnvBool("ALLOW_PARTIAL_SUCCESS", true)
	config.Server.Upload.EnableBatchProcessing = getEnvBool("ENABLE_BATCH_PROCESSING", true)
	config.Server.Upload.BatchSize = getEnvInt("BATCH_SIZE", 10)
	config.Server.Upload.MaxConcurrentUploads = getEnvInt("MAX_CONCURRENT_UPLOADS", 3)
	config.Server.Upload.StreamingThreshold = getEnvInt64("STREAMING_THRESHOLD", 10485760)
	config.Server.Upload.ValidateBeforeUpload = getEnvBool("VALIDATE_BEFORE_UPLOAD", true)
	config.Server.Upload.EnableProgressTracking = getEnvBool("ENABLE_PROGRESS_TRACKING", false)
	config.Server.Upload.CleanupOnFailure = getEnvBool("CLEANUP_ON_FAILURE", false)
	config.Server.Upload.RateLimitPerMinute = getEnvInt("RATE_LIMIT_PER_MINUTE", 100)

	// Database configuration
	config.Database.DSN = getEnvString("DB_DSN", "./data/cloudlet.db")
	config.Database.MaxConn = getEnvInt("DB_MAX_CONN", 10)

	return nil
}

// loadFromYAML loads configuration from YAML file (fallback method)
func loadFromYAML(config *Config, cfgPath string) error {
	file, err := os.Open(cfgPath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	return decoder.Decode(config)
}

// getEnvString returns the environment variable value or default if not set
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns the environment variable as int or default if not set/invalid
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvInt64 returns the environment variable as int64 or default if not set/invalid
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
		}
	}
	return defaultValue
}

// getEnvBool returns the environment variable as bool or default if not set/invalid
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		// Handle various boolean representations
		lowerValue := strings.ToLower(value)
		switch lowerValue {
		case "true", "1", "yes", "on", "enabled":
			return true
		case "false", "0", "no", "off", "disabled":
			return false
		}
	}
	return defaultValue
}

// hasAnyEnvVars checks if any of the configuration environment variables are set
func hasAnyEnvVars() bool {
	envVars := []string{
		"PORT",
		"MAX_MEMORY",
		"MAX_FILE_SIZE",
		"STORAGE_PATH",
		"READ_TIMEOUT",
		"WRITE_TIMEOUT",
		"IDLE_TIMEOUT",
		"MAX_HEADER_BYTES",
		"READ_HEADER_TIMEOUT",
		"MAX_FILES_PER_REQUEST",
		"MAX_TOTAL_SIZE_PER_REQUEST",
		"ALLOW_PARTIAL_SUCCESS",
		"ENABLE_BATCH_PROCESSING",
		"BATCH_SIZE",
		"MAX_CONCURRENT_UPLOADS",
		"STREAMING_THRESHOLD",
		"VALIDATE_BEFORE_UPLOAD",
		"ENABLE_PROGRESS_TRACKING",
		"CLEANUP_ON_FAILURE",
		"RATE_LIMIT_PER_MINUTE",
		"DB_DSN",
		"DB_MAX_CONN",
	}

	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}
	return false
}
