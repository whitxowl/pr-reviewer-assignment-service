package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/app"
	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/config"
)

const (
	shutdownTimeout = 10 * time.Second
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("application panic", "err", r)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cfg := config.MustLoadConfig()

	log := config.NewLogger(cfg.Env)
	slog.SetDefault(log)

	log.Info("starting pr-reviewer application")

	application := app.New(ctx, log, cfg)

	go application.Srv.MustRun(ctx)

	<-ctx.Done()

	log.Info("received stop signal")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := application.Srv.Stop(shutdownCtx)
	if err != nil {
		log.Error("failed to stop http_server gracefully", "err", err)
	} else {
		log.Info("PR Reviewer application http_server stopped gracefully")
	}
}
