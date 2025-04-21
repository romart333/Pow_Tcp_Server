package main

import (
	"Pow_Tcp_Server/internal/app/config"
	"Pow_Tcp_Server/internal/app/repository/inmemory_repo"
	"Pow_Tcp_Server/internal/app/services"
	"Pow_Tcp_Server/internal/app/transport/tcpserver"
	"context"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Error("Error while flushing logger", zap.Error(err))
		}
	}(logger)

	// Load configuration
	cfg, err := config.ReadServerConfig()
	if err != nil {
		logger.Fatal("Failed to read config", zap.Error(err))
	}

	// Initialize services
	quoteRepo := inmemory_repo.NewInMemoryQuoteRepo()
	quoteService := services.NewQuoteService(quoteRepo)
	powService := services.NewPOWService(cfg.POWDifficulty)

	// Create TCP server
	server := tcpserver.NewServer(
		cfg,
		tcpserver.NewHandler(
			powService,
			quoteService,
			logger,
			cfg.ReadTimeout,
			cfg.WriteTimeout,
			cfg.POWCalcTimeout,
		),
		logger,
	)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Run(ctx); err != nil {
			logger.Error("Server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		server.Stop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Server stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn("Forced shutdown due to timeout")
	}
}
