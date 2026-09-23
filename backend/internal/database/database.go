package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(context context.Context, databaseUrl string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context, databaseUrl)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
