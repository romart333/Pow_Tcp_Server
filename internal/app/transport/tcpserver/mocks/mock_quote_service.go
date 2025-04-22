package mocks

import "github.com/romart333/Pow_Tcp_Server/internal/app/domain"

type MockQuoteService struct {
	GetRandomQuoteFn func() (*domain.Quote, error)
}

func (m *MockQuoteService) GetRandomQuote() (*domain.Quote, error) {
	return m.GetRandomQuoteFn()
}
