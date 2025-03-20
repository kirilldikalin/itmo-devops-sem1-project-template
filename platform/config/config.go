package config

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	DefaultDateFormat        = "2006-01-02"
	DefaultReadTimeout       = 1 * time.Minute
	DefaultWriteTimeout      = 1 * time.Minute
	DefaultMaxFileSize       = 10 * 1024 * 1024 // 10 MB
	DefaultFileUploadTimeout = 1 * time.Minute
	DefaultConfigName        = "config"
	DefaultConfigType        = "yaml"
)

type ServerSettings struct {
	Port              int           `mapstructure:"port"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	FileUploadTimeout time.Duration
	MaxFileSize       int64
}

type DatabaseSettings struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type Settings struct {
	Server   ServerSettings   `mapstructure:"server"`
	Database DatabaseSettings `mapstructure:"database"`
}

func parseFileSize(value string) (int64, error) {
	if strings.HasSuffix(value, "KB") {
		num, err := strconv.ParseInt(strings.TrimSuffix(value, "KB"), 10, 64)
		if err != nil {
			return 0, err
		}
		return num * 1024, nil
	} else if strings.HasSuffix(value, "MB") {
		num, err := strconv.ParseInt(strings.TrimSuffix(value, "MB"), 10, 64)
		if err != nil {
			return 0, err
		}
		return num * 1024 * 1024, nil
	} else if strings.HasSuffix(value, "GB") {
		num, err := strconv.ParseInt(strings.TrimSuffix(value, "GB"), 10, 64)
		if err != nil {
			return 0, err
		}
		return num * 1024 * 1024 * 1024, nil
	} else {
		return strconv.ParseInt(value, 10, 64)
	}
}

func validateConfig(cfg *Settings) error {
	if cfg.Server.ReadTimeout < 5*time.Second {
		return fmt.Errorf("read_timeout must be at least 5 seconds")
	}
	if cfg.Server.WriteTimeout < 5*time.Second {
		return fmt.Errorf("write_timeout must be at least 5 seconds")
	}
	if cfg.Server.MaxFileSize <= 0 {
		return fmt.Errorf("max_file_size must be greater than 0")
	}
	if cfg.Server.FileUploadTimeout < 5*time.Second {
		return fmt.Errorf("file_upload_timeout must be at least 5 seconds")
	}
	return nil
}

func Load(configPath string) (Settings, error) {
	log.Printf("Loading configuration from '%s'\n", configPath)

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName(DefaultConfigName)
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}

	viper.SetConfigType(DefaultConfigType)

	if err := viper.ReadInConfig(); err != nil {
		return Settings{}, fmt.Errorf("failed to read configuration: %w", err)
	}

	var cfg Settings
	if err := viper.Unmarshal(&cfg); err != nil {
		return Settings{}, fmt.Errorf("failed to parse configuration: %w", err)
	}

	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = DefaultReadTimeout
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = DefaultWriteTimeout
	}
	if cfg.Server.FileUploadTimeout == 0 {
		cfg.Server.FileUploadTimeout = DefaultFileUploadTimeout
	}

	rawSize := viper.GetString("server.max_file_size")
	log.Printf("Raw max_file_size value: %s", rawSize)

	if rawSize == "" {
		cfg.Server.MaxFileSize = DefaultMaxFileSize
		log.Printf("Using default max_file_size: %d bytes", DefaultMaxFileSize)
	} else {
		parsedSize, err := parseFileSize(rawSize)
		if err != nil {
			return Settings{}, fmt.Errorf("invalid max_file_size value: %w", err)
		}
		cfg.Server.MaxFileSize = parsedSize
		log.Printf("Parsed max_file_size value: %d bytes", cfg.Server.MaxFileSize)
	}

	if err := validateConfig(&cfg); err != nil {
		return Settings{}, fmt.Errorf("invalid configuration: %w", err)
	}

	log.Println("Configuration loaded successfully")
	return cfg, nil
}