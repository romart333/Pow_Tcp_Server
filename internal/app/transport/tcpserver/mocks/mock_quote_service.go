package mocks

import "Pow_Tcp_Server/internal/app/domain"

type MockQuoteService struct {
	GetRandomQuoteFn func() (*domain.Quote, error)
}

func (m *MockQuoteService) GetRandomQuote() (*domain.Quote, error) {
	return m.GetRandomQuoteFn()
}
