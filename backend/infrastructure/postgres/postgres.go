package postgres

import (
	"context"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"vfinancy/backend/infrastructure/database"
	"vfinancy/backend/infrastructure/logger"
)

const DriverName = "pgx"

func ConnectDSN(ctx context.Context, dsn string, log *logger.Logger) (*database.DB, error) {
	log.Info("connecting to postgres")

	db, err := database.Open(DriverName, dsn, database.Options{
		MaxOpenConns: 4,
		MaxIdleConns: 2,
	})
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	log.Info("postgres connection established")
	return db, nil
}


