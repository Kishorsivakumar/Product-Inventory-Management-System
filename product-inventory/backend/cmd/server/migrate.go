package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migrate runs the idempotent SQL file at startup.
func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	sql, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, string(sql))
	return err
}
