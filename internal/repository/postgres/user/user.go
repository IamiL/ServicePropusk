package postgresUser

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (*Storage, error) {
	return &Storage{db: pool}, nil
}

func (s *Storage) NewUser(
	ctx context.Context,
	login string,
	passHash string,
) error {
	_, err := s.db.Exec(
		ctx,
		"INSERT INTO users(login, pass_hash) VALUES($1, $2)",
		login,
		passHash,
	)
	if err != nil {
		return err
	}

	return nil
}
