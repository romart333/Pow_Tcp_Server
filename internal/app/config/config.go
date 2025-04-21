package config

import (
	"Pow_Tcp_Server/internal/app/domain"
	"Pow_Tcp_Server/internal/common"
	"fmt"
	"os"
	"strconv"
	"time"
)

// ServerConfig defines all server configuration parameters
type ServerConfig struct {
	Address         string        `yaml:"address"`          // TCP address to listen on
	POWDifficulty   int           `yaml:"pow_difficulty"`   // Proof-of-Work difficulty level (1-10)
	POWCalcTimeout  time.Duration `yaml:"pow_calc_timeout"` // Timeout for POW calculation
	ReadTimeout     time.Duration `yaml:"read_timeout"`     // Timeout for read operations
	WriteTimeout    time.Duration `yaml:"write_timeout"`    // Timeout for write operations
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"` // Graceful shutdown timeout
	MaxConns        int           `yaml:"max_connections"`  // Maximum concurrent connections
	AcceptTimeout   time.Duration `yaml:"accept_timeout"`   // Timeout for accepting new connections
}

// ClientConfig defines all client configuration parameters
type ClientConfig struct {
	ServerAddress  string        `yaml:"server_address"`   // Server address to connect to
	POWTimeout     time.Duration `yaml:"pow_timeout"`      // Timeout for POW calculation
	DialTimeout    time.Duration `yaml:"dial_timeout"`     // Timeout for connection establishment
	MaxRetries     int           `yaml:"max_retries"`      // Maximum number of retry attempts
	BaseRetryDelay time.Duration `yaml:"base_retry_delay"` // Initial delay between retries
	MaxRetryDelay  time.Duration `yaml:"max_retry_delay"`  // Maximum delay between retries
}

// ReadServerConfig loads server configuration from environment variables
// with fallback to default values. Validates POW difficulty range.
func ReadServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{
		Address:         common.DefaultServerPort,
		POWDifficulty:   common.DefaultPOWDifficulty,
		POWCalcTimeout:  common.DefaultPOWCalcTimeout,
		ReadTimeout:     common.DefaultReadTimeout,
		WriteTimeout:    common.DefaultWriteTimeout,
		ShutdownTimeout: common.DefaultShutdownTimeout,
		MaxConns:        common.DefaultMaxConns,
		AcceptTimeout:   common.DefaultAcceptTimeout,
	}

	// Load configuration from environment variables
	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		cfg.Address = addr
	}

	if conns := os.Getenv("MAX_CONNECTIONS"); conns != "" {
		maxConnections, err := strconv.Atoi(conns)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_CONNECTIONS value: %w", err)
		}
		cfg.MaxConns = maxConnections
	}

	if diffStr := os.Getenv("POW_DIFFICULTY"); diffStr != "" {
		diff, err := strconv.Atoi(diffStr)
		if err != nil {
			return nil, fmt.Errorf("invalid POW_DIFFICULTY value: %w", err)
		}
		if diff < common.MinPOWDifficulty || diff > common.MaxPOWDifficulty {
			return nil, domain.ErrInvalidDifficulty
		}
		cfg.POWDifficulty = diff
	}

	// Load timeout values
	cfg.POWCalcTimeout = loadDuration("POW_CALC_TIMEOUT", common.DefaultPOWCalcTimeout)
	cfg.ReadTimeout = loadDuration("READ_TIMEOUT", common.DefaultReadTimeout)
	cfg.WriteTimeout = loadDuration("WRITE_TIMEOUT", common.DefaultWriteTimeout)
	cfg.ShutdownTimeout = loadDuration("SHUTDOWN_TIMEOUT", common.DefaultShutdownTimeout)
	cfg.AcceptTimeout = loadDuration("ACCEPT_TIMEOUT", common.DefaultAcceptTimeout)

	return cfg, nil
}

// ReadClientConfig loads client configuration from environment variables
// with fallback to default values.
func ReadClientConfig() (*ClientConfig, error) {
	cfg := &ClientConfig{
		ServerAddress:  common.DefaultClientTarget,
		POWTimeout:     common.DefaultPOWTimeout,
		DialTimeout:    common.DefaultDialTimeout,
		MaxRetries:     common.DefaultMaxRetries,
		BaseRetryDelay: common.DefaultBaseRetryDelay,
		MaxRetryDelay:  common.DefaultMaxRetryDelay,
	}

	if addr := os.Getenv("CLIENT_SERVER_ADDRESS"); addr != "" {
		cfg.ServerAddress = addr
	}

	// Load timeout values
	cfg.POWTimeout = loadDuration("CLIENT_POW_TIMEOUT", common.DefaultPOWTimeout)
	cfg.DialTimeout = loadDuration("CLIENT_DIAL_TIMEOUT", common.DefaultDialTimeout)

	if retriesStr := os.Getenv("CLIENT_MAX_RETRIES"); retriesStr != "" {
		if retries, err := strconv.Atoi(retriesStr); err == nil {
			cfg.MaxRetries = retries
		}
	}

	cfg.BaseRetryDelay = loadDuration("CLIENT_BASE_RETRY_DELAY", common.DefaultBaseRetryDelay)
	cfg.MaxRetryDelay = loadDuration("CLIENT_MAX_RETRY_DELAY", common.DefaultMaxRetryDelay)

	return cfg, nil
}

// loadDuration helper function to load duration from environment variable
func loadDuration(envVar string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(envVar); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
