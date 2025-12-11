package main

import (
	"log/slog"

	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/config"
)

func main() {
	cfg := config.MustLoadConfig()
	log := config.NewLogger(cfg.Env)
	log.Info("Starting pr-reviewer-assignment-service", slog.String("env", cfg.Env))

}
