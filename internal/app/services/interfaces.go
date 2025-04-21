package services

import "Pow_Tcp_Server/internal/app/domain"

type QuoteRepository interface {
	GetRandom() (*domain.Quote, error)
}
