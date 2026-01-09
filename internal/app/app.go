package app

import (
	"context"
	"log/slog"

	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/app/server"
	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/config"
	teamService "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/team"
	userService "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/user"
	teamStorage "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/storage/team"
	userStorage "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/storage/user"
	"github.com/whitxowl/pr-reviewer-assignment-service.git/pkg/postgres"
)

type App struct {
	Srv *server.Server
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) *App {
	pgPool, err := postgres.NewPool(ctx, cfg.StorageConfig.DSN(), postgres.WithMaxConnections(int32(cfg.StorageConfig.MaxConnections)))
	if err != nil {
		panic("failed to connect to database" + err.Error())
	}

	teamStore := teamStorage.New(pgPool)
	userStore := userStorage.New(pgPool)

	teamSvc := teamService.New(log.WithGroup("service.team"), teamStore, userStore)
	userSvc := userService.New(log.WithGroup("service.user"), userStore)

	srv := server.New(log, teamSvc, userSvc, cfg.HTTPServer)

	return &App{
		Srv: srv,
	}
}
