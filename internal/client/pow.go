package client

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"github.com/romart333/Pow_Tcp_Server/internal/common"
	"time"
)

// solvePOW solves the PoW challenge by finding the correct nonce.
func solvePOW(challenge []byte, difficulty int, timeout time.Duration) (int, error) {
	deadline := time.Now().Add(timeout)

	for nonce := 0; ; nonce++ {
		if time.Now().After(deadline) {
			return 0, errors.New("POW solving timeout")
		}

		// Verify the solution by hashing the challenge and nonce
		if verifySolution(challenge, difficulty, nonce) {
			return nonce, nil
		}
	}
}

// verifySolution checks if the given nonce solves the challenge by matching the difficulty criteria.
func verifySolution(challenge []byte, difficulty, nonce int) bool {
	hashInput := make([]byte, common.HashInputSize)
	copy(hashInput[:common.ChallengeDataSize], challenge)
	binary.LittleEndian.PutUint64(hashInput[common.ChallengeDataSize:], uint64(nonce))

	hash := sha256.Sum256(hashInput)

	// Check if the hash meets the difficulty requirements
	zeroBytes := difficulty / common.BitsPerByte
	remainingBits := difficulty % common.BitsPerByte

	for i := 0; i < zeroBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}

	if remainingBits > 0 {
		mask := byte(0xFF) << (common.BitsPerByte - remainingBits)
		return (hash[zeroBytes] & mask) == 0
	}

	return true
}
