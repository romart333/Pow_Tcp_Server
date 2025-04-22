package main

import (
	"context"
	"github.com/romart333/Pow_Tcp_Server/internal/app/config"
	"github.com/romart333/Pow_Tcp_Server/internal/client"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
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
	cfg, err := config.ReadClientConfig()
	if err != nil {
		logger.Fatal("Failed to read config", zap.Error(err))
	}

	// Create client
	cli := client.NewClient(cfg, logger)

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Handle SIGINT/SIGTERM
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	// Get quote
	quote, err := cli.GetQuote(ctx)
	if err != nil {
		logger.Error("Failed to get quote", zap.Error(err))
		return
	}

	logger.Info("Successfully received quote", zap.String("quote", quote))
}
