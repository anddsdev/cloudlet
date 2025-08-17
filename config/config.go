package config

import (
	"os"

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

func NewConfig(cfgPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
