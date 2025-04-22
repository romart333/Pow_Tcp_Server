package client

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/romart333/Pow_Tcp_Server/internal/app/config"
	"github.com/romart333/Pow_Tcp_Server/internal/common"
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
		c.logger.Info("Attempting to get quote", zap.Int("attempt", i+1), zap.Duration("retry_delay", retryDelay))
		quote, err := c.tryGetQuote(ctx)
		if err == nil {
			c.logger.Info("Successfully received quote", zap.String("quote", quote))
			return quote, nil
		}
		lastErr = err

		c.logger.Warn("Attempt failed",
			zap.Int("attempt", i+1),
			zap.Error(err),
			zap.Duration("retry_delay", retryDelay))

		select {
		case <-ctx.Done():
			c.logger.Error("Context cancelled", zap.Error(ctx.Err()))
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
	c.logger.Info("Dialing server", zap.String("server_address", c.config.ServerAddress))
	dialer := &net.Dialer{Timeout: c.config.DialTimeout}
	conn, err := dialer.DialContext(ctx, common.DefaultNetwork, c.config.ServerAddress)
	if err != nil {
		c.logger.Error("Connection failed", zap.Error(err), zap.String("server_address", c.config.ServerAddress))
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			c.logger.Error("Closing connection failed", zap.Error(err))
		} else {
			c.logger.Info("Connection closed successfully", zap.String("server_address", c.config.ServerAddress))
		}
	}(conn)

	c.logger.Info("Setting deadline for connection", zap.Duration("deadline", c.config.POWTimeout))
	deadline := time.Now().Add(c.config.POWTimeout)
	if err := conn.SetDeadline(deadline); err != nil {
		c.logger.Error("Set deadline failed", zap.Error(err))
		return "", fmt.Errorf("set deadline failed: %w", err)
	}

	challenge, difficulty, err := c.receiveChallenge(conn)
	if err != nil {
		c.logger.Error("Failed to receive challenge", zap.Error(err))
		return "", fmt.Errorf("receive challenge failed: %w", err)
	}

	c.logger.Info("Solving POW challenge", zap.Int("difficulty", difficulty))
	nonce, err := solvePOW(challenge, difficulty, c.config.POWTimeout)
	if err != nil {
		c.logger.Error("Failed to solve POW challenge", zap.Error(err))
		return "", fmt.Errorf("solve POW failed: %w", err)
	}

	c.logger.Info("Sending POW solution", zap.Int("nonce", nonce))
	if err := c.sendSolution(conn, nonce); err != nil {
		c.logger.Error("Failed to send solution", zap.Error(err))
		return "", fmt.Errorf("send solution failed: %w", err)
	}

	c.logger.Info("Receiving quote from server")
	return c.receiveQuote(conn)
}

func (c *TCPClient) receiveChallenge(conn net.Conn) ([]byte, int, error) {
	c.logger.Info("Receiving challenge data")
	challenge := make([]byte, common.ChallengeDataSize)
	if _, err := io.ReadFull(conn, challenge); err != nil {
		c.logger.Error("Failed to read challenge data", zap.Error(err))
		return nil, 0, fmt.Errorf("read challenge data failed: %w", err)
	}

	var difficulty int
	if err := binary.Read(conn, binary.BigEndian, &difficulty); err != nil {
		c.logger.Error("Failed to read difficulty", zap.Error(err))
		return nil, 0, fmt.Errorf("read difficulty failed: %w", err)
	}

	c.logger.Info("Received challenge and difficulty", zap.Int("difficulty", difficulty))
	return challenge, difficulty, nil
}

func (c *TCPClient) sendSolution(conn net.Conn, nonce int) error {
	c.logger.Info("Sending solution", zap.Int("nonce", nonce))
	return binary.Write(conn, binary.BigEndian, int64(nonce))
}

func (c *TCPClient) receiveQuote(conn net.Conn) (string, error) {
	c.logger.Info("Receiving quote from server")
	buf := make([]byte, common.ResponseBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		c.logger.Error("Failed to read quote", zap.Error(err))
		return "", fmt.Errorf("read quote failed: %w", err)
	}

	response := string(buf[:n])
	c.logger.Info("Received server response", zap.String("response", response))

	if len(response) >= common.MinPrefixLength {
		switch {
		case response[:len(common.ErrorPrefix)] == common.ErrorPrefix:
			c.logger.Error("Server returned error", zap.String("error", response[len(common.ErrorPrefix):]))
			return "", errors.New(response[len(common.ErrorPrefix):])
		case response[:len(common.QuotePrefix)] == common.QuotePrefix:
			c.logger.Info("Received valid quote", zap.String("quote", response[len(common.QuotePrefix):]))
			return response[len(common.QuotePrefix):], nil
		}
	}

	c.logger.Error("Invalid server response format", zap.String("response", response))
	return "", errors.New("invalid server response format")
}
