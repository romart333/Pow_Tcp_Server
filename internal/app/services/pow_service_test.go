package services

import (
	"Pow_Tcp_Server/internal/app/domain"
	"testing"
)

func TestNewPOWService(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int
		wantErr    bool
	}{
		{"valid difficulty", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewPOWService(tt.difficulty)

			if service == nil {
				t.Fatal("Expected non-nil POWService, got nil")
			}

			if service.difficulty != tt.difficulty {
				t.Errorf("Expected difficulty %d, got %d", tt.difficulty, service.difficulty)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	// create test challenge with known parameters
	testChallenge := &domain.POWChallenge{
		Data:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Difficulty: 12, // 12 leading zero bits (1.5 bytes)
	}

	service := NewPOWService(12) // difficulty should match the test challenge

	tests := []struct {
		name     string
		nonce    int
		expected bool
	}{
		{"valid nonce", 12345, true},
		{"invalid nonce", 0, false},
		{"negative nonce", -1, false},
		{"large nonce", 999999999, false},
	}

	// Find valid nonce for testing challenge
	validNonceFound := false
	for nonce := 0; nonce < 100000; nonce++ {
		if service.Verify(testChallenge, nonce) {
			tests[0].nonce = nonce
			validNonceFound = true
			break
		}
	}

	if !validNonceFound {
		t.Fatal("Could not find valid nonce for testing")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := service.Verify(testChallenge, tt.nonce); got != tt.expected {
				t.Errorf("Verify() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVerifyWithDifferentDifficulties(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int
		nonce      int
		expected   bool
	}{
		{"low difficulty valid", 4, 42, true},
		{"high difficulty invalid", 20, 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create testing challenge
			challenge := &domain.POWChallenge{
				Data:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				Difficulty: tt.difficulty,
			}

			service := NewPOWService(tt.difficulty)

			// for test with low difficulty find actual nonce
			if tt.expected {
				found := false
				for nonce := 0; nonce < 100000; nonce++ {
					if service.Verify(challenge, nonce) {
						found = true
						break
					}
				}
				if !found {
					t.Fatal("Could not find valid nonce for low difficulty")
				}
				// use found nonce
				tt.nonce = findValidNonce(service, challenge)
			}

			if got := service.Verify(challenge, tt.nonce); got != tt.expected {
				t.Errorf("Verify() with difficulty %d = %v, want %v", tt.difficulty, got, tt.expected)
			}
		})
	}
}

// helper to search for valid nonce
func findValidNonce(service *POWService, challenge *domain.POWChallenge) int {
	for nonce := 0; nonce < 100000; nonce++ {
		if service.Verify(challenge, nonce) {
			return nonce
		}
	}
	return 0
}
