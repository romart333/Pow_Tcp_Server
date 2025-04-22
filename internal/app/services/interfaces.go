package services

import "github.com/romart333/Pow_Tcp_Server/internal/app/domain"

type QuoteRepository interface {
	GetRandom() (*domain.Quote, error)
}
