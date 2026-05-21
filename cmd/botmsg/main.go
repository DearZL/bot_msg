package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bot_msg/internal/config"
	"bot_msg/internal/event"
	"bot_msg/internal/server"
	"bot_msg/internal/service"
	"bot_msg/pkg/onebot"
)

// @title Bot Msg OneBot API
// @version 0.1.0
// @description Bot Msg can run as a OneBot v11 HTTP proxy service backed by a OneBot HTTP/HTTPS or WS/WSS endpoint.
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	cfg.OneBot.Logger = logger
	oneBotClient := onebot.NewClientWithConfig(cfg.OneBot)
	eventDispatcher := event.NewDispatcher(logger)
	eventDispatcher.Register(oneBotClient)

	oneBotCtx, stopOneBot := context.WithCancel(context.Background())
	defer stopOneBot()
	oneBotClient.Start(oneBotCtx)

	msgService := service.New(cfg.Bot, oneBotClient)

	httpServer := server.New(cfg.HTTP, msgService, logger)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("onebot http api server started", "addr", cfg.HTTP.Addr)
		errCh <- httpServer.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}
	stopOneBot()
	logger.Info("server shutdown completed")
}
