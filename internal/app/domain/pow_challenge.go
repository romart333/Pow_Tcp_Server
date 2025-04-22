package domain

import (
	"crypto/rand"
	"github.com/romart333/Pow_Tcp_Server/internal/common"
)

// POWChallenge represents the structure of a Proof of Work challenge
type POWChallenge struct {
	Data       [common.ChallengeDataSize]byte
	Difficulty int
}

// NewPOWChallenge creates a new POW challenge with random data and configurable difficulty
func NewPOWChallenge(difficulty int) (*POWChallenge, error) {
	var data [common.ChallengeDataSize]byte
	if _, err := rand.Read(data[:]); err != nil {
		return nil, err
	}

	return &POWChallenge{
		Data:       data,
		Difficulty: difficulty,
	}, nil
}
