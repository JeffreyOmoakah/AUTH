package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/JeffreyOmoakah/AUTH.git/internal/env"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	_ "github.com/lib/pq"
)



func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	port := env.GetString("PORT", "3000")
	cfg := config{
		addr: ":" + port,
		db: dbConfig{
			dsn: env.GetString("DATABASE_URL", ""), 
		},
	}
	
	// Logger first
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	
	logger.Info("starting migrations...")
	
	// Run migrations with sql.DB
	sqlDB, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		logger.Error("failed to open DB for migrations", "error", err)
		os.Exit(1)
	}
	
	// Test connection with timeout
	ctx2, cancel2 := context.WithTimeout(ctx, 10*time.Second)
	defer cancel2()
	if err := sqlDB.PingContext(ctx2); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	logger.Info("database connection successful")
	
	// Set provider for PostgreSQL
	goose.SetDialect("postgres")
	
	if err := goose.Up(sqlDB, "./internal/adapters/postgresql/migrations"); err != nil {
		logger.Error("migration failed", "error", err)
		sqlDB.Close()
		os.Exit(1)
	}
	sqlDB.Close()
	logger.Info("migrations completed successfully")
	
	// Now connect with pgxpool
	pool, err := pgxpool.New(ctx, cfg.db.dsn)
	if err != nil {
		logger.Error("unable to connect to database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("database connection pool established")
	
	api := application{
		config: cfg,
		db:     pool,
		logger: logger,
	}
	
	logger.Info("server starting", "addr", cfg.addr)
	if err := api.run(api.mount()); err != nil {
		logger.Error("server crashed", "error", err)
		os.Exit(1)
	}
}