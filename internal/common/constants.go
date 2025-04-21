package common

import "time"

const (
	BitsPerByte = 8
)

// Network constants
const (
	DefaultNetwork      = "tcp"
	DefaultServerPort   = ":8080"
	DefaultClientTarget = "localhost" + DefaultServerPort
	DefaultMaxConns     = 1000
)

// POW constants
const (
	ChallengeDataSize    = 16 // Size of POW challenge data in bytes (128-bit)
	HashInputSize        = 24 // ChallengeDataSize + NonceSize
	MinPOWDifficulty     = 1  // Minimum difficulty level
	MaxPOWDifficulty     = 10 // Maximum difficulty level
	DefaultPOWDifficulty = 3  // Default difficulty level
	ResponseBufferSize   = 1024
)

// Timeout constants
const (
	DefaultPOWCalcTimeout  = 25 * time.Second
	DefaultReadTimeout     = 5 * time.Second
	DefaultWriteTimeout    = 5 * time.Second
	DefaultShutdownTimeout = 10 * time.Second
	DefaultPOWTimeout      = 30 * time.Second
	DefaultDialTimeout     = 3 * time.Second
	DefaultAcceptTimeout   = 500 * time.Millisecond
)

// Retry constants
const (
	DefaultMaxRetries     = 3
	DefaultBaseRetryDelay = 1 * time.Second
	DefaultMaxRetryDelay  = 10 * time.Second
	MaxBackoffJitterRatio = 0.25 // 25% of retry delay
)

// Protocol constants
const (
	ErrorPrefix       = "ERROR:"
	QuotePrefix       = "QUOTE:"
	ResponseDelimiter = "|"
	MinPrefixLength   = 6 // len("ERROR:") or len("QUOTE:")
)
