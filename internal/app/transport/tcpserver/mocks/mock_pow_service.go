package mocks

import "github.com/romart333/Pow_Tcp_Server/internal/app/domain"

type MockPOWService struct {
	GenerateChallengeFn func() (*domain.POWChallenge, error)
	VerifyFn            func(challenge *domain.POWChallenge, nonce int) bool
}

func (m *MockPOWService) GenerateChallenge() (*domain.POWChallenge, error) {
	return m.GenerateChallengeFn()
}

func (m *MockPOWService) Verify(challenge *domain.POWChallenge, nonce int) bool {
	return m.VerifyFn(challenge, nonce)
}
