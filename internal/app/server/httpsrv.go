package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	teamHandler "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/api/v1/team"
	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/config"
	teamService "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/team"
)

type Server struct {
	log         *slog.Logger
	teamService *teamService.Service
	cfg         *config.HTTPServer

	mu     sync.Mutex
	server *http.Server
}

func New(log *slog.Logger, teamService *teamService.Service, cfg config.HTTPServer) *Server {
	return &Server{
		log:         log,
		teamService: teamService,
		cfg:         &cfg,
	}
}

func (s *Server) MustRun(ctx context.Context) {
	if err := s.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic("failed to run HTTP server: " + err.Error())
	}
}

func (s *Server) Run(ctx context.Context) error {
	const op = "httpapp.Run"

	log := s.log.With(
		slog.String("op", op),
		slog.String("address", s.cfg.Address),
	)

	log.InfoContext(ctx, "starting HTTP server")

	teamHdlr := teamHandler.New(s.teamService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(ginLogger(s.log))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	base := router.Group("/")

	teamHdlr.RegisterRoutes(base)

	srv := &http.Server{
		Addr:         s.cfg.Address,
		Handler:      router,
		ReadTimeout:  s.cfg.Timeout,
		WriteTimeout: s.cfg.Timeout,
		IdleTimeout:  s.cfg.IdleTimeout,
	}

	s.mu.Lock()
	s.server = srv
	s.mu.Unlock()

	log.InfoContext(ctx, "HTTP server started successfully")

	return srv.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	const op = "httpapp.Stop"

	log := s.log.With(slog.String("op", op))

	log.Info("stopping HTTP server")

	s.mu.Lock()
	server := s.server
	s.mu.Unlock()

	if server == nil {
		return nil
	}

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("HTTP server stopped successfully")

	return nil
}

func ginLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		log.Info("HTTP request",
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
			slog.String("user_agent", c.Request.UserAgent()),
		)
	}
}
