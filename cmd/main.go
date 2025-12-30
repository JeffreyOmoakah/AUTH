package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/JeffreyOmoakah/AUTH.git/internal/env"
	"github.com/jackc/pgx/v5"
)



func main() {
	ctx := context.Background()
	
	cfg := config{
		addr: env.GetString("ADDR", ":3000"),
		db: dbConfig{
			dsn:env.GetString("GOOSE_DBSTRING", "host=localhost user=postgres password=postgres dbname=auth sslmode=disable"),
		},
	}
	
	// Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	
	conn, err := pgx.Connect(ctx, cfg.db.dsn)
		if err != nil {
			slog.Error("failed to connect to database", "error", err)
			os.Exit(1)
		}
		defer conn.Close(ctx)
		
		logger.Info("connected to database","dsn",cfg.db.dsn)
		
		api := application{
			config: cfg,
			db: conn,
		}
		
		api.run(api.mount())
		
			if err := api.run(api.mount()); err != nil {
				slog.Error("server failed to start", "error", err)
				os.Exit(1)
			}
}