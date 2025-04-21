package client

import (
	"Pow_Tcp_Server/internal/app/config"
	"Pow_Tcp_Server/internal/common"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"time"

	"go.uber.org/zap"
)

type TCPClient struct {
	config *config.ClientConfig
	logger *zap.Logger
}

func NewClient(cfg *config.ClientConfig, logger *zap.Logger) *TCPClient {
	return &TCPClient{
		config: cfg,
		logger: logger,
	}
}

func (c *TCPClient) GetQuote(ctx context.Context) (string, error) {
	var lastErr error
	retryDelay := c.config.BaseRetryDelay

	for i := 0; i < c.config.MaxRetries; i++ {
		quote, err := c.tryGetQuote(ctx)
		if err == nil {
			return quote, nil
		}
		lastErr = err

		c.logger.Warn("Attempt failed",
			zap.Int("attempt", i+1),
			zap.Error(err),
			zap.Duration("retry_delay", retryDelay))

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(retryDelay):
			retryDelay = time.Duration(math.Min(
				float64(c.config.MaxRetryDelay),
				float64(retryDelay*2),
			))
			jitter := time.Duration(rand.Int63n(int64(float64(retryDelay) * common.MaxBackoffJitterRatio)))
			retryDelay += jitter
		}
	}

	return "", fmt.Errorf("after %d attempts: %w", c.config.MaxRetries, lastErr)
}

func (c *TCPClient) tryGetQuote(ctx context.Context) (string, error) {
	dialer := &net.Dialer{Timeout: c.config.DialTimeout}
	conn, err := dialer.DialContext(ctx, common.DefaultNetwork, c.config.ServerAddress)
	if err != nil {
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			c.logger.Error("Closing connection failed", zap.Error(err))
		}
	}(conn)

	deadline := time.Now().Add(c.config.POWTimeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return "", fmt.Errorf("set deadline failed: %w", err)
	}

	challenge, difficulty, err := c.receiveChallenge(conn)
	if err != nil {
		return "", fmt.Errorf("receive challenge failed: %w", err)
	}

	nonce, err := solvePOW(challenge, difficulty, c.config.POWTimeout)
	if err != nil {
		return "", fmt.Errorf("solve POW failed: %w", err)
	}

	if err := c.sendSolution(conn, nonce); err != nil {
		return "", fmt.Errorf("send solution failed: %w", err)
	}

	return c.receiveQuote(conn)
}

func (c *TCPClient) receiveChallenge(conn net.Conn) ([]byte, int, error) {
	challenge := make([]byte, common.ChallengeDataSize)
	if _, err := io.ReadFull(conn, challenge); err != nil {
		return nil, 0, fmt.Errorf("read challenge data failed: %w", err)
	}

	var difficulty int
	if err := binary.Read(conn, binary.BigEndian, &difficulty); err != nil {
		return nil, 0, fmt.Errorf("read difficulty failed: %w", err)
	}

	return challenge, difficulty, nil
}

func (c *TCPClient) sendSolution(conn net.Conn, nonce int) error {
	return binary.Write(conn, binary.BigEndian, int64(nonce))
}

func (c *TCPClient) receiveQuote(conn net.Conn) (string, error) {
	buf := make([]byte, common.ResponseBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("read quote failed: %w", err)
	}

	response := string(buf[:n])
	if len(response) >= common.MinPrefixLength {
		switch {
		case response[:len(common.ErrorPrefix)] == common.ErrorPrefix:
			return "", errors.New(response[len(common.ErrorPrefix):])
		case response[:len(common.QuotePrefix)] == common.QuotePrefix:
			return response[len(common.QuotePrefix):], nil
		}
	}

	return "", errors.New("invalid server response format")
}
