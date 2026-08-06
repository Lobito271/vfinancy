package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// DB wraps *sql.DB and adds a WithTx helper. The driver name is passed
// by the caller (e.g. "pgx" via _ "github.com/jackc/pgx/v5/stdlib" or
// "postgres" via _ "github.com/lib/pq").
type DB struct {
	*sql.DB
}

type Tx struct {
	*sql.Tx
}

type Options struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Open(driverName, dsn string, opts Options) (*DB, error) {
	if dsn == "" {
		return nil, errors.New("database: empty DSN")
	}
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}
	if opts.MaxOpenConns > 0 {
		db.SetMaxOpenConns(opts.MaxOpenConns)
	}
	if opts.MaxIdleConns > 0 {
		db.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(opts.ConnMaxLifetime)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}
	return &DB{DB: db}, nil
}

func (d *DB) Close() error {
	return d.DB.Close()
}

func (d *DB) WithTx(ctx context.Context, fn func(tx *Tx) error) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("database: begin: %w", err)
	}
	t := &Tx{Tx: tx}
	if err := fn(t); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			return fmt.Errorf("database: rollback after error %v: %w", err, rbErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("database: commit: %w", err)
	}
	return nil
}
