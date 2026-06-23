package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/kirillshakirov/mcp-task-server/internal/mcp"
	"github.com/kirillshakirov/mcp-task-server/internal/tools"
)

type config struct {
	Addr        string `env:"ADDR"         env-default:":8080"`
	DatabaseURL string `env:"DATABASE_URL"  env-default:"postgres://tasks:tasks@localhost:5432/tasks"`
	RedisURL    string `env:"REDIS_URL"     env-default:"redis://localhost:6379"`
	LogLevel    string `env:"LOG_LEVEL"     env-default:"debug"`
	APIKey      string `env:"MCP_API_KEY"    env-required:"true"`
	ReqApiKey   bool   `env:"REQ_API_KEY"    env-required:"true"`
}

func main() {
	var cfg config
	// cleanenv reads .env first, then real env vars override.
	// If .env doesn't exist that's fine — it moves on silently.
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		// File missing is not an error; only hard failures abort.
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			slog.Error("config", "err", err)
			os.Exit(1)
		}
	}

	level := slog.LevelDebug
	if cfg.LogLevel == "info" {
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))

	ctx := context.Background()

	// Postgres
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("pgxpool.New", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		slog.Error("postgres ping failed", "err", err)
		os.Exit(1)
	}
	slog.Info("postgres connected")

	// Redis
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("redis.ParseURL", "err", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("redis ping failed", "err", err)
		os.Exit(1)
	}
	slog.Info("redis connected")

	// MCP server
	registry := mcp.NewRegistry()
	tools.RegisterAll(registry, db, rdb)
	srv := mcp.NewServer("mcp-task-server", "0.1.0", cfg.APIKey, registry, cfg.ReqApiKey)

	httpSrv := &http.Server{
		Addr:        cfg.Addr,
		Handler:     srv.Handler(),
		ReadTimeout: 10 * time.Second,
		// WriteTimeout intentionally omitted — SSE connections are long-lived.
		IdleTimeout: 120 * time.Second,
	}

	go func() {
		slog.Info("mcp-task-server listening", "addr", cfg.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
}
