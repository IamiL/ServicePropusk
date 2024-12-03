package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnPool() (*pgxpool.Pool, error) {
	return pgxpool.New(
		context.Background(),
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			"localhost",
			"5432",
			"iamil-admin",
			"adminpass",
			"service-propusk",
		),
	)
}
