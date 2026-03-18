package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-arch/internal/app"
	"go-arch/internal/infra/configuration"
	infralog "go-arch/internal/infra/log"
)

func main() {
	mode := flag.String("mode", "http", "Run mode: http, cron, nsq")
	configTest := flag.Bool("t", false, "Test configuration and exit")
	configPath := flag.String("config", "files/config/app.yaml", "Path to config file")
	flag.Parse()

	// Init logger
	logger := infralog.NewLogger()

	// Load config
	cfg, err := configuration.Load(*configPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	if *configTest {
		logger.Info().Msg("configuration is valid")
		os.Exit(0)
	}

	// Init application (skip infra connections for now — they'll fail without real DBs)
	application := app.New(cfg, logger, nil, nil)
	defer application.Close()

	// Context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	switch *mode {
	case "http":
		runHTTP(ctx, application, sigCh)
	case "cron":
		runCron(ctx, application, sigCh)
	case "nsq":
		runNSQ(ctx, application, sigCh)
	default:
		logger.Fatal().Str("mode", *mode).Msg("unknown mode")
	}
}

func runHTTP(ctx context.Context, application *app.Application, sigCh chan os.Signal) {
	addr := fmt.Sprintf(":%d", application.Config.Server.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      application.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		application.Logger.Info().Str("addr", addr).Msg("http server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			application.Logger.Fatal().Err(err).Msg("http server error")
		}
	}()

	<-sigCh
	application.Logger.Info().Msg("shutting down http server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		application.Logger.Error().Err(err).Msg("http server shutdown error")
	}

	application.Logger.Info().Msg("http server stopped")
}

func runCron(ctx context.Context, application *app.Application, sigCh chan os.Signal) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-sigCh
		application.Logger.Info().Msg("shutting down cron...")
		cancel()
	}()

	if err := application.CronHandler.Start(ctx); err != nil {
		application.Logger.Error().Err(err).Msg("cron error")
	}
}

func runNSQ(ctx context.Context, application *app.Application, sigCh chan os.Signal) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-sigCh
		application.Logger.Info().Msg("shutting down nsq consumer...")
		cancel()
	}()

	if err := application.NSQHandler.Start(ctx); err != nil {
		application.Logger.Error().Err(err).Msg("nsq consumer error")
	}
}
