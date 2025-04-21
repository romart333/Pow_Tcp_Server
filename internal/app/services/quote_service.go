package services

import (
	"Pow_Tcp_Server/internal/app/domain"
)

type QuoteService struct {
	repo QuoteRepository
}

func NewQuoteService(repo QuoteRepository) *QuoteService {
	return &QuoteService{repo: repo}
}

func (s *QuoteService) GetRandomQuote() (*domain.Quote, error) {
	return s.repo.GetRandom()
}
