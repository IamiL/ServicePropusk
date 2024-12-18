package postgresBuilds

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	model "rip/internal/domain"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (*Storage, error) {
	return &Storage{db: pool}, nil
}

func (s *Storage) Buildings(ctx context.Context) (
	[]model.BuildingModel,
	error,
) {
	const op = "repository.services.postgres.Buildings"

	query := `SELECT id, name, description, img_url FROM buildings WHERE status = 'true'`

	rows, err := s.db.Query(context.TODO(), query)
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	services := []model.BuildingModel{}
	for rows.Next() {
		c := model.BuildingModel{}
		err := rows.Scan(
			&c.Id,
			&c.Name,
			&c.Description,
			&c.ImgUrl,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: execute statement: %w", op, err)
		}
		services = append(services, c)
	}

	return services, nil
}

func (s *Storage) Building(ctx context.Context, id string) (
	model.BuildingModel,
	error,
) {
	const op = "repository.services.postgres.Build"

	query := `SELECT name, description, img_url FROM buildings WHERE status = 'true' AND id = $1`

	var build model.BuildingModel

	build.Id = id

	err := s.db.QueryRow(ctx, query, id).Scan(
		&build.Name,
		&build.Description,
		&build.ImgUrl,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BuildingModel{}, errors.New("build not found")
		}
		return model.BuildingModel{}, fmt.Errorf(
			"%s: execute statement: %w",
			op,
			err,
		)
	}
	return build, nil
}

func (s *Storage) EditImgUrl(
	ctx context.Context,
	id string,
	url string,
) error {
	const op = "repository.services.postgres.EditImgUrl"

	query := `UPDATE buildings SET img_url = $1 WHERE id = $2;`

	_, err := s.db.Exec(ctx, query, url, id)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return nil

}

func (s *Storage) EditBuildingInfo(
	ctx context.Context,
	building *model.BuildingModel,
) error {
	const op = "repository.services.postgres.EditBuilding"

	query := `UPDATE buildings SET name = $1, description = $2 WHERE id = $3;`

	_, err := s.db.Exec(
		ctx,
		query,
		building.Name,
		building.Description,
		building.Id,
	)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return nil
}

func (s *Storage) EditBuildingStatus(
	ctx context.Context,
	id string,
	status bool,
) error {
	const op = "repository.services.postgres.EditBuilding"

	query := `UPDATE buildings SET status = $1 WHERE id = $2;`

	_, err := s.db.Exec(
		ctx,
		query,
		status,
		id,
	)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return nil
}
