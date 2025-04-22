package tcpserver

import (
	"context"
	"errors"
	"github.com/romart333/Pow_Tcp_Server/internal/app/config"
	"go.uber.org/zap"
	"net"
	"sync"
)

// Server represents a TCP server instance
type Server struct {
	cfg         *config.ServerConfig
	handler     *Handler
	logger      *zap.Logger
	listener    net.Listener
	connections map[net.Conn]struct{}
	connMutex   sync.RWMutex
	shutdown    chan struct{}
	wg          sync.WaitGroup
	connLimit   chan struct{} // Connection limiter semaphore
}

// NewServer creates a new TCP server instance
func NewServer(cfg *config.ServerConfig, handler *Handler, logger *zap.Logger) *Server {
	return &Server{
		cfg:         cfg,
		handler:     handler,
		logger:      logger,
		connections: make(map[net.Conn]struct{}),
		shutdown:    make(chan struct{}),
		connLimit:   make(chan struct{}, cfg.MaxConns),
	}
}

// Run starts the server and listens for incoming connections
func (s *Server) Run(ctx context.Context) error {
	var err error
	s.listener, err = net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return err
	}

	s.logger.Info("Server started",
		zap.String("address", s.cfg.Address),
		zap.Int("max_connections", s.cfg.MaxConns))

	// Handle shutdown signals
	go func() {
		<-ctx.Done()
		s.logger.Info("Shutdown signal received")
		close(s.shutdown)
		err := s.listener.Close()
		if err != nil {
			s.logger.Error("Error while closing listener", zap.Error(err))
		}
	}()

	for {
		select {
		case <-s.shutdown:
			return nil
		case s.connLimit <- struct{}{}: // Acquire connection slot
			conn, err := s.acceptWithTimeout(ctx)
			if err != nil {
				<-s.connLimit // Release slot on error
				if errors.Is(err, net.ErrClosed) {
					return err // Listener closed
				}
				s.logger.Warn("Failed to accept connection", zap.Error(err))
				continue
			}

			s.trackConnection(conn)
			s.wg.Add(1)
			go func() {
				defer func() { <-s.connLimit }() // Release slot when done
				s.handleConnection(ctx, conn)
			}()
		}
	}
}

// acceptWithTimeout adds timeout for Accept operation
func (s *Server) acceptWithTimeout(ctx context.Context) (net.Conn, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan net.Conn)
	errChan := make(chan error, 1)

	go func() {
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Error("Error accepting connection", zap.Error(err))
			errChan <- err
			return
		}
		s.logger.Info("Connection accepted", zap.String("remote_addr", conn.RemoteAddr().String()))
		done <- conn
	}()

	select {
	case conn := <-done:
		return conn, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() {
	s.logger.Info("Closing active connections",
		zap.Int("count", len(s.connections)))

	s.connMutex.Lock()
	for conn := range s.connections {
		if err := conn.Close(); err != nil {
			s.logger.Warn("Failed to close connection",
				zap.Error(err),
				zap.String("remote_addr", conn.RemoteAddr().String()))
		}
	}
	s.connMutex.Unlock()

	s.wg.Wait()

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.Warn("Failed to close listener", zap.Error(err))
		}
	}

	s.logger.Info("Server stopped gracefully")
}

// trackConnection adds connection to active connections map
func (s *Server) trackConnection(conn net.Conn) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	s.connections[conn] = struct{}{}
}

// untrackConnection removes connection from active connections map
func (s *Server) untrackConnection(conn net.Conn) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	delete(s.connections, conn)
}

// handleConnection processes a single client connection
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer s.wg.Done()
	defer s.untrackConnection(conn)
	defer func() {
		if err := conn.Close(); err != nil && !isClosedError(err) {
			s.logger.Warn("Connection close error",
				zap.Error(err),
				zap.String("remote_addr", conn.RemoteAddr().String()))
		}
	}()

	select {
	case <-s.shutdown:
		return
	default:
		s.handler.Handle(ctx, conn)
	}
}

// isClosedError checks if error is related to closed connection
func isClosedError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, net.ErrClosed)
}
