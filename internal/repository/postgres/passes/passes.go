package postgresPasses

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	model "rip/internal/domain"
	"time"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (*Storage, error) {
	return &Storage{db: pool}, nil
}

func (s *Storage) Pass(ctx context.Context, id string) (
	*model.PassModel,
	error,
) {
	const op = "repository.passes.postgres.Pass"

	query := `SELECT visitor, visit_date FROM passes WHERE id = $1 AND status = 0`

	pass := model.PassModel{}

	err := s.db.QueryRow(ctx, query, id).Scan(
		&pass.VisitorName,
		&pass.DateVisit,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("error errnorows")
			return &model.PassModel{}, errors.New("pass not found")
		}
		fmt.Println("error other")
		return &model.PassModel{}, fmt.Errorf(
			"%s: execute statement: %w",
			op,
			err,
		)
	}

	fmt.Println("getPass repo 1")

	query = `SELECT b.id, b.name, b.description, b.img_url, bs.comment FROM buildings_passes bs JOIN buildings b ON bs.building = b.id WHERE pass = $1`

	rows, err := s.db.Query(ctx, query, id)
	if err != nil {
		return &model.PassModel{}, fmt.Errorf(
			"%s: execute statement: %w",
			op,
			err,
		)
	}
	defer rows.Close()

	fmt.Println("getPass repo 2")

	pass.ID = id

	pass.Items = make(model.PassItems, 0, 5)

	fmt.Println("getPass repo 3")

	for rows.Next() {

		fmt.Println("getPass repo 4 start")

		var c pgtype.Text

		item := model.PassItem{Building: &model.BuildingModel{}}

		err := rows.Scan(
			&item.Building.Id,
			&item.Building.Name,
			&item.Building.Description,
			&item.Building.ImgUrl,
			&c,
		)
		if err != nil {
			return &model.PassModel{}, fmt.Errorf(
				"%s: execute statement: %w",
				op,
				err,
			)
		}

		fmt.Println("getPass repo 4 seredina")

		item.Comment = c.String

		pass.Items = append(pass.Items, &item)

		fmt.Println("getPass repo 4 end")
	}

	return &pass, nil
}

func (s *Storage) ID(ctx context.Context, uid string) (string, error) {
	const op = "repository.passes.postgres.Pass"

	var id string

	query := `SELECT p.id
				FROM passes AS p
				WHERE p.creator = $1 AND
				      status = 0;`

	err := s.db.QueryRow(ctx, query, uid).Scan(
		&id,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("error errnorows")
			return "", nil
		}
		fmt.Println("error other")
		return "", fmt.Errorf(
			"%s: execute statement: %w",
			op,
			err,
		)
	}

	return id, nil
}

func (s *Storage) AddToPass(
	ctx context.Context,
	recordID string,
	id string,
	buildingID string,
) error {
	const op = "repository.passes.postgres.Pass"

	query := `INSERT INTO buildings_passes (id, building, pass) VALUES ($1, $2, $3);`

	_, err := s.db.Exec(ctx, query, recordID, buildingID, id)
	if err != nil {
		fmt.Println("ошибка: ", err.Error())
		return fmt.Errorf("unable to insert row: %w", err)
	}

	fmt.Println("успешно")
	return nil
}

func (s *Storage) Delete(ctx context.Context, id string) error {
	query := `UPDATE passes SET status = 1 WHERE id = $1;`

	_, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return nil
}

func (s *Storage) NewDraftPass(
	ctx context.Context,
	id string,
	uid string,
	visitor string,
	visitDate time.Time,
) error {
	query := `INSERT INTO passes (id, creator, creation_date, visitor, visit_date, status) VALUES ($1, $2, $3, $4, $5, 0);`

	if _, err := s.db.Exec(
		ctx,
		query,
		id,
		uid,
		time.Now(),
		visitor,
		visitDate,
	); err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return nil
}

func (s *Storage) DraftPassIDByCreator(
	ctx context.Context,
	uid string,
) (string, error) {
	const op = "repository.passes.postgres.Pass"

	var id string

	query := `SELECT id FROM passes WHERE creator = $1 AND status = 0`

	err := s.db.QueryRow(ctx, query, uid).Scan(
		&id,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("error errnorows")
			return "", errors.New("pass not found")
		}
		fmt.Println("error other")
		return "", fmt.Errorf(
			"%s: execute statement: %w",
			op,
			err,
		)
	}

	return id, nil
}

func (s *Storage) ItemsCount(ctx context.Context, uid string) (
	int,
	error,
) {
	const op = "repository.passes.postgres.Pass"

	var count int

	query := `SELECT COUNT(bp.*)
				FROM passes AS p
				LEFT JOIN buildings_passes AS bp ON p.id = bp.pass
				WHERE p.creator = $1 AND status = 0
				GROUP BY p.creator;`

	err := s.db.QueryRow(ctx, query, uid).Scan(
		&count,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("error errnorows")
			return 0, errors.New("pass not found")
		}
		fmt.Println("error other")
		return 0, fmt.Errorf(
			"%s: execute statement: %w",
			op,
			err,
		)
	}

	return count, nil
}
