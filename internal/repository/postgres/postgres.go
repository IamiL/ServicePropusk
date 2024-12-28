package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func NewConnPool() (*pgx.Conn, error) {
	return pgx.Connect(
		context.Background(),
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			"localhost",
			"5430",
			"iamil-admin",
			"adminpass",
			"service-propusk",
		),
	)
}
