package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(
		context.Background(),
		connString,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to database: %w",
			err,
		)
	}

	return pool, nil
}