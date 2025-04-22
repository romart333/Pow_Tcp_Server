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
	h.logger.Info("Handling new connection")

	// Set initial read deadline for challenge request
	if err := conn.SetReadDeadline(time.Now().Add(h.readTimeout)); err != nil {
		h.logger.Error("Failed to set read deadline", zap.Error(err))
		return
	}

	// 1. Generate challenge
	h.logger.Info("Generating POW challenge")
	challenge, err := h.powService.GenerateChallenge()
	if err != nil {
		h.sendError(conn, "Failed to generate challenge", err)
		return
	}

	// 2. Send challenge
	h.logger.Info("Sending POW challenge")
	if err := h.sendChallenge(conn, challenge); err != nil {
		h.logger.Error("Failed to send challenge", zap.Error(err))
		return
	}

	// 3. Receive nonce
	h.logger.Info("Waiting for nonce from client")
	nonce, err := h.receiveNonce(conn)
	if err != nil {
		h.sendError(conn, "Invalid nonce", err)
		return
	}
	h.logger.Info("Received nonce from client", zap.Int("nonce", nonce))

	// 4. Verify nonce
	h.logger.Info("Verifying POW nonce")
	if !h.powService.Verify(challenge, nonce) {
		h.sendError(conn, "Invalid proof", nil)
		return
	}
	h.logger.Info("Nonce verified successfully")

	// 5. Send quote
	h.logger.Info("Sending quote to client")
	if err := h.sendQuote(conn); err != nil {
		h.logger.Error("Failed to send quote", zap.Error(err))
	}
}

// Private helper methods
func (h *Handler) sendChallenge(conn net.Conn, challenge *domain.POWChallenge) error {
	h.logger.Info("Sending challenge data")

	if err := conn.SetWriteDeadline(time.Now().Add(h.writeTimeout)); err != nil {
		h.logger.Error("Failed to set write deadline for challenge", zap.Error(err))
		return err
	}

	// Write challenge data (16 bytes)
	if _, err := conn.Write(challenge.Data[:]); err != nil {
		h.logger.Error("Failed to write challenge data", zap.Error(err))
		return err
	}

	// Convert difficulty to int32 before writing (fixed size)
	difficulty := int32(challenge.Difficulty)
	if err := binary.Write(conn, binary.BigEndian, difficulty); err != nil {
		h.logger.Error("Failed to write challenge difficulty", zap.Error(err))
		return err
	}
	return nil
}

func (h *Handler) receiveNonce(conn net.Conn) (int, error) {
	h.logger.Info("Receiving nonce from client")

	// Extended timeout for POW calculation
	if err := conn.SetReadDeadline(time.Now().Add(h.powTimeout)); err != nil {
		h.logger.Error("Failed to set read deadline for nonce", zap.Error(err))
		return 0, err
	}

	var nonce int64
	if err := binary.Read(conn, binary.BigEndian, &nonce); err != nil {
		h.logger.Error("Failed to read nonce", zap.Error(err))
		return 0, err
	}

	h.logger.Info("Successfully received nonce", zap.Int64("nonce", nonce))
	return int(nonce), nil
}

func (h *Handler) sendQuote(conn net.Conn) error {
	h.logger.Info("Sending quote to client")

	if err := conn.SetWriteDeadline(time.Now().Add(h.writeTimeout)); err != nil {
		h.logger.Error("Failed to set write deadline for quote", zap.Error(err))
		return err
	}

	quote, err := h.quoteService.GetRandomQuote()
	if err != nil {
		h.logger.Error("Failed to get random quote", zap.Error(err))
		return err
	}

	_, err = conn.Write([]byte("QUOTE:" + quote.Text + "|" + quote.Author))
	if err != nil {
		h.logger.Error("Failed to send quote", zap.Error(err))
	}
	return err
}

func (h *Handler) sendError(conn net.Conn, msg string, err error) {
	h.logger.Error(msg, zap.Error(err))

	err = conn.SetWriteDeadline(time.Now().Add(h.writeTimeout))
	if err != nil {
		h.logger.Error("Failed to set write deadline for error", zap.Error(err))
		return
	}
	_, err = conn.Write([]byte("ERROR:" + msg))
	if err != nil {
		h.logger.Error("Failed to send error", zap.Error(err))
	}
}
