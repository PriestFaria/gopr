// @title PR Reviewer Assignment Service
// @version 1.0
// @BasePath /api/v1

package main

import (
	"context"
	"errors"
	"gopr/cmd/config"
	"gopr/internal/gateways/rest"
	"gopr/internal/usecase"
	"gopr/pkg/slogx"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load(".env")
	cfg.Print()

	log := cfg.Logger()
	slog.SetDefault(log)
	log.Info("Hello from gopr server!")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx = slogx.NewCtx(ctx, log)

	pgConfig := cfg.PGXConfig()
	pool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		log.Error("can't create new database pool", slogx.Err(err))
		os.Exit(1)
	}
	defer pool.Close()

	s := rest.NewServer(ctx, cfg, usecase.Setup(ctx, cfg, pool))
	if err := s.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slogx.WithErr(log, err).Error("error during server shutdown")
	}
}
