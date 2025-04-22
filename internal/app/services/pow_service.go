package services

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/romart333/Pow_Tcp_Server/internal/app/domain"
	"github.com/romart333/Pow_Tcp_Server/internal/common"
)

type POWService struct {
	difficulty int
}

// NewPOWService creates a new Proof-of-Work service with configurable difficulty
func NewPOWService(difficulty int) *POWService {
	return &POWService{
		difficulty: difficulty,
	}
}

// GenerateChallenge creates a new POW challenge with the configured difficulty
func (s *POWService) GenerateChallenge() (*domain.POWChallenge, error) {
	challenge, err := domain.NewPOWChallenge(s.difficulty)
	if err != nil {
		return nil, err
	}
	return challenge, nil
}

// Verify checks if the nonce solves the challenge
func (s *POWService) Verify(challenge *domain.POWChallenge, nonce int) bool {
	// Prepare input data: challenge data + nonce
	hashInput := make([]byte, common.HashInputSize)
	copy(hashInput[:common.ChallengeDataSize], challenge.Data[:])
	binary.LittleEndian.PutUint64(hashInput[common.ChallengeDataSize:], uint64(nonce)) // Add nonce in the last 8 bytes

	// Calculate the SHA-256 hash of the input
	hash := sha256.Sum256(hashInput)

	// Check for the required number of leading zeros based on difficulty
	requiredZeros := challenge.Difficulty
	zeroBytes := requiredZeros / common.BitsPerByte
	remainingBits := requiredZeros % common.BitsPerByte

	// Verify the full zero bytes
	for i := 0; i < zeroBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}

	// Check remaining bits if needed
	if remainingBits > 0 {
		mask := byte(0xFF) << (common.BitsPerByte - remainingBits)
		if (hash[zeroBytes] & mask) != 0 {
			return false
		}
	}

	return true
}
