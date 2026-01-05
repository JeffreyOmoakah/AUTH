package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"

	"github.com/JeffreyOmoakah/AUTH.git/internal/env"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
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
	
	// Run migrations
	sqlDB, err := sql.Open("pgx", cfg.db.dsn)
	if err != nil {
		log.Fatalf("Failed to open DB for migrations: %v", err)
	}
	
	if err := goose.Up(sqlDB, "./internal/adapters/postgresql/migrations"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	sqlDB.Close()
	log.Println("Migrations completed successfully")
	
	// Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)
	
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

			// 4. Proper blocking start
			logger.Info("server starting", "addr", cfg.addr)
			if err := api.run(api.mount()); err != nil {
				logger.Error("server crashed", "error", err)
				os.Exit(1)
			}
}