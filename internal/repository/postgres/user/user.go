package postgresUser

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type Storage struct {
	db *pgx.Conn
}

func New(pool *pgx.Conn) (*Storage, error) {
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

func (s *Storage) User(ctx context.Context, login string) (
	string,
	string,
	error,
) {
	query := `SELECT id, pass_hash FROM users WHERE login = $1`

	var id, passHash string

	err := s.db.QueryRow(ctx, query, login).Scan(&id, &passHash)
	if err != nil {
		return "", "", err
	}

	return id, passHash, nil
}

func (s *Storage) EditUser(
	ctx context.Context,
	uid string,
	login string,
	passHash string,
) error {
	query := `UPDATE users SET login = &1, pass_hash = $2 WHERE id = $3`

	_, err := s.db.Exec(ctx, query, login, passHash, uid)
	if err != nil {
		return err
	}

	return nil
}
