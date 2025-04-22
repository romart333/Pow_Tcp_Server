package tcpserver

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"github.com/romart333/Pow_Tcp_Server/internal/app/domain"
	"github.com/romart333/Pow_Tcp_Server/internal/app/transport/tcpserver/mocks"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

// TestHandler_Handle tests the main connection handling logic
func TestHandler_Handle(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	timeouts := time.Second * 1

	tests := []struct {
		name         string
		setupMocks   func() (*mocks.MockPOWService, *mocks.MockQuoteService)
		prepareConn  func() *mocks.TestConn
		wantErr      bool
		expectOutput string
	}{
		{
			name: "successful flow",
			setupMocks: func() (*mocks.MockPOWService, *mocks.MockQuoteService) {
				pow := &mocks.MockPOWService{
					GenerateChallengeFn: func() (*domain.POWChallenge, error) {
						return &domain.POWChallenge{
							Data:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
							Difficulty: 4,
						}, nil
					},
					VerifyFn: func(challenge *domain.POWChallenge, nonce int) bool {
						return true
					},
				}

				quote := &mocks.MockQuoteService{
					GetRandomQuoteFn: func() (*domain.Quote, error) {
						return &domain.Quote{
							Text:   "Test quote",
							Author: "Test author",
						}, nil
					},
				}

				return pow, quote
			},
			prepareConn: func() *mocks.TestConn {
				conn := &mocks.TestConn{}
				// Prepare nonce response (int64 = 8 bytes)
				var nonce int64 = 42
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uint64(nonce))
				conn.ReadBuf = buf
				return conn
			},
			wantErr:      false,
			expectOutput: "QUOTE:Test quote|Test author",
		},
		{
			name: "challenge generation error",
			setupMocks: func() (*mocks.MockPOWService, *mocks.MockQuoteService) {
				pow := &mocks.MockPOWService{
					GenerateChallengeFn: func() (*domain.POWChallenge, error) {
						return nil, errors.New("generation error")
					},
				}
				return pow, &mocks.MockQuoteService{}
			},
			prepareConn:  func() *mocks.TestConn { return &mocks.TestConn{} },
			wantErr:      true,
			expectOutput: "ERROR:Failed to generate challenge",
		},
		{
			name: "invalid nonce",
			setupMocks: func() (*mocks.MockPOWService, *mocks.MockQuoteService) {
				pow := &mocks.MockPOWService{
					GenerateChallengeFn: func() (*domain.POWChallenge, error) {
						return &domain.POWChallenge{
							Data:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
							Difficulty: 4,
						}, nil
					},
					VerifyFn: func(challenge *domain.POWChallenge, nonce int) bool {
						return false
					},
				}
				return pow, &mocks.MockQuoteService{}
			},
			prepareConn: func() *mocks.TestConn {
				conn := &mocks.TestConn{}
				var nonce int64 = 42
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uint64(nonce))
				conn.ReadBuf = buf
				return conn
			},
			wantErr:      true,
			expectOutput: "ERROR:Invalid proof",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pow, quote := tt.setupMocks()
			conn := tt.prepareConn()

			handler := NewHandler(
				pow,
				quote,
				logger,
				timeouts,
				timeouts,
				timeouts,
			)

			handler.Handle(ctx, conn)

			if tt.wantErr {
				if len(conn.WriteBuf) == 0 {
					t.Error("Expected error response, got nothing")
				}
			}

			if tt.expectOutput != "" {
				output := string(conn.WriteBuf)
				if !strings.Contains(output, tt.expectOutput) {
					t.Errorf("Expected output to contain %q, got %q", tt.expectOutput, output)
				}
			}
		})
	}
}

// TestHandler_sendChallenge tests the challenge sending functionality
func TestHandler_sendChallenge(t *testing.T) {
	logger := zaptest.NewLogger(t)
	conn := &mocks.TestConn{}

	handler := &Handler{
		logger:       logger,
		writeTimeout: 1 * time.Second,
	}

	challenge := &domain.POWChallenge{
		Data:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Difficulty: 4,
	}

	err := handler.sendChallenge(conn, challenge)
	if err != nil {
		t.Fatalf("sendChallenge failed: %v", err)
	}

	// Verify total bytes written (16 data + 4 difficulty)
	if len(conn.WriteBuf) != 20 {
		t.Errorf("Expected 20 bytes written, got %d", len(conn.WriteBuf))
	}

	// Verify challenge data portion
	if !bytes.Equal(conn.WriteBuf[:16], challenge.Data[:]) {
		t.Error("Challenge data portion doesn't match")
	}

	// Verify difficulty portion (int32 big endian)
	var difficulty int32
	buf := bytes.NewReader(conn.WriteBuf[16:20])
	if err := binary.Read(buf, binary.BigEndian, &difficulty); err != nil {
		t.Fatalf("Failed to read difficulty: %v", err)
	}
	if difficulty != 4 {
		t.Errorf("Expected difficulty 4, got %d", difficulty)
	}
}

// TestHandler_receiveNonce tests the nonce receiving functionality
func TestHandler_receiveNonce(t *testing.T) {
	logger := zaptest.NewLogger(t)
	conn := &mocks.TestConn{}

	// Prepare test data - nonce 42 (int64 = 8 bytes)
	var nonce int64 = 42
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(nonce))
	conn.ReadBuf = buf

	handler := &Handler{
		logger:     logger,
		powTimeout: time.Second,
	}

	result, err := handler.receiveNonce(conn)
	if err != nil {
		t.Fatalf("receiveNonce failed: %v", err)
	}

	if result != 42 {
		t.Errorf("Expected nonce 42, got %d", result)
	}
}

// TestHandler_sendQuote tests the quote sending functionality
func TestHandler_sendQuote(t *testing.T) {
	logger := zaptest.NewLogger(t)
	conn := &mocks.TestConn{}

	handler := &Handler{
		logger:       logger,
		writeTimeout: time.Second,
		quoteService: &mocks.MockQuoteService{
			GetRandomQuoteFn: func() (*domain.Quote, error) {
				return &domain.Quote{
					Text:   "Test quote",
					Author: "Test author",
				}, nil
			},
		},
	}

	err := handler.sendQuote(conn)
	if err != nil {
		t.Fatalf("sendQuote failed: %v", err)
	}

	expected := "QUOTE:Test quote|Test author"
	if string(conn.WriteBuf) != expected {
		t.Errorf("Expected %q, got %q", expected, string(conn.WriteBuf))
	}
}

// TestHandler_sendError tests the error sending functionality
func TestHandler_sendError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	conn := &mocks.TestConn{}

	handler := &Handler{
		logger:       logger,
		writeTimeout: time.Second,
	}

	handler.sendError(conn, "test error", errors.New("test"))

	expected := "ERROR:test error"
	if !strings.Contains(string(conn.WriteBuf), expected) {
		t.Errorf("Expected output to contain %q, got %q", expected, string(conn.WriteBuf))
	}
}
