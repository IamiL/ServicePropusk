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
	uid string,
	login string,
	passHash string,
	isAdmin bool,
) error {
	_, err := s.db.Exec(
		ctx,
		"INSERT INTO users(id, login, pass_hash, is_admin) VALUES($1, $2, $3, $4)",
		uid,
		login,
		passHash,
		isAdmin,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) User(ctx context.Context, login string) (
	string,
	bool,
	string,
	error,
) {
	query := `SELECT id, is_admin, pass_hash FROM users WHERE login = $1`

	var id, passHash string
	var isAdmin bool

	err := s.db.QueryRow(ctx, query, login).Scan(&id, &isAdmin, &passHash)
	if err != nil {
		return "", false, "", err
	}

	return id, isAdmin, passHash, nil
}

func (s *Storage) EditUser(
	ctx context.Context,
	uid string,
	login string,
	passHash string,
) error {
	query := `UPDATE users SET login = $1, pass_hash = $2 WHERE id = $3`

	_, err := s.db.Exec(ctx, query, login, passHash, uid)
	if err != nil {
		return err
	}

	return nil
}
