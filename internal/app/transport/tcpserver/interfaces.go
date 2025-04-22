package tcpserver

import (
	"github.com/romart333/Pow_Tcp_Server/internal/app/domain"
)

type POWService interface {
	GenerateChallenge() (*domain.POWChallenge, error)
	Verify(challenge *domain.POWChallenge, nonce int) bool
}

type QuoteService interface {
	GetRandomQuote() (*domain.Quote, error)
}
