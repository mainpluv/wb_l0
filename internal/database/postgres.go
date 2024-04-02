package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(dataSource string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dataSource)
	if err != nil {
		log.Fatalf("Error creating postgres pool: %v", err)
	}
	return pool, nil
}
