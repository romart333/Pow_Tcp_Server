package tcpserver

import (
	"context"
	"encoding/binary"
	"github.com/romart333/Pow_Tcp_Server/internal/app/domain"
	"net"
	"time"

	"go.uber.org/zap"
)

// Handler processes incoming client connections
type Handler struct {
	powService   POWService
	quoteService QuoteService
	logger       *zap.Logger
	readTimeout  time.Duration
	writeTimeout time.Duration
	powTimeout   time.Duration
}

// NewHandler creates a new connection handler
func NewHandler(
	pow POWService,
	quote QuoteService,
	logger *zap.Logger,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	powTimeout time.Duration,
) *Handler {
	return &Handler{
		powService:   pow,
		quoteService: quote,
		logger:       logger,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		powTimeout:   powTimeout,
	}
}

// Handle processes a client connection
func (h *Handler) Handle(ctx context.Context, conn net.Conn) {
	// Set initial read deadline for challenge request
	if err := conn.SetReadDeadline(time.Now().Add(h.readTimeout)); err != nil {
		h.logger.Error("Failed to set read deadline", zap.Error(err))
		return
	}

	// 1. Generate challenge
	challenge, err := h.powService.GenerateChallenge()
	if err != nil {
		h.sendError(conn, "Failed to generate challenge", err)
		return
	}

	// 2. send challenge
	if err := h.sendChallenge(conn, challenge); err != nil {
		h.logger.Error("Failed to send challenge", zap.Error(err))
		return
	}

	// 3. Receive nonce
	nonce, err := h.receiveNonce(conn)
	if err != nil {
		h.sendError(conn, "Invalid nonce", err)
		return
	}

	// 4. Verify nonce
	if !h.powService.Verify(challenge, nonce) {
		h.sendError(conn, "Invalid proof", nil)
		return
	}

	// 5. Send quote
	if err := h.sendQuote(conn); err != nil {
		h.logger.Error("Failed to send quote", zap.Error(err))
	}
}

// Private helper methods
func (h *Handler) sendChallenge(conn net.Conn, challenge *domain.POWChallenge) error {
	if err := conn.SetWriteDeadline(time.Now().Add(h.writeTimeout)); err != nil {
		return err
	}

	// Write challenge data (16 bytes)
	if _, err := conn.Write(challenge.Data[:]); err != nil {
		return err
	}

	// Convert difficulty to int32 before writing (fixed size)
	difficulty := int32(challenge.Difficulty)
	return binary.Write(conn, binary.BigEndian, difficulty)
}

func (h *Handler) receiveNonce(conn net.Conn) (int, error) {
	// Extended timeout for POW calculation
	if err := conn.SetReadDeadline(time.Now().Add(h.powTimeout)); err != nil {
		return 0, err
	}

	var nonce int64
	if err := binary.Read(conn, binary.BigEndian, &nonce); err != nil {
		return 0, err
	}

	return int(nonce), nil
}

func (h *Handler) sendQuote(conn net.Conn) error {
	if err := conn.SetWriteDeadline(time.Now().Add(h.writeTimeout)); err != nil {
		return err
	}

	quote, err := h.quoteService.GetRandomQuote()
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte("QUOTE:" + quote.Text + "|" + quote.Author))
	return err
}

func (h *Handler) sendError(conn net.Conn, msg string, err error) {
	h.logger.Error(msg,
		zap.String("remote_addr", conn.RemoteAddr().String()),
		zap.Error(err))

	err = conn.SetWriteDeadline(time.Now().Add(h.writeTimeout))
	if err != nil {
		h.logger.Error("Failed to set write deadline", zap.Error(err))
		return
	}
	_, err = conn.Write([]byte("ERROR:" + msg))
	if err != nil {
		h.logger.Error("Failed to send error", zap.Error(err))
	}
}
