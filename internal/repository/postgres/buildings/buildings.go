package postgresBuildings

import (
	"context"
	"errors"
	"fmt"
	model "service-propusk-backend/internal/domain"
	repoErrors "service-propusk-backend/internal/pkg/errors/repo"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (*Storage, error) {
	return &Storage{db: pool}, nil
}

func (s *Storage) AllBuildings(ctx context.Context) (
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

func (s *Storage) FindBuildings(ctx context.Context, name string) (
	[]model.BuildingModel,
	error,
) {
	const op = "repository.services.postgres.Buildings"

	query := `SELECT id, name, description, img_url FROM buildings WHERE status = 'true' AND name LIKE '%` + name + `%' `

	rows, err := s.db.Query(context.TODO(), query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repoErrors.ErrorNotFound
		}

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
			return model.BuildingModel{}, repoErrors.ErrorNotFound
		}
		return model.BuildingModel{}, fmt.Errorf(
			"%s: execute statement: %w",
			op,
			err,
		)
	}
	return build, nil
}

func (s *Storage) EditBuildingImgUrl(
	ctx context.Context,
	id string,
	url string,
) error {
	const op = "repository.services.postgres.EditBuildingImgUrl"

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

	query := `UPDATE buildings SET name = $1, description = $2 WHERE id = $3 AND status = 'true';`

	result, err := s.db.Exec(
		ctx,
		query,
		building.Name,
		building.Description,
		building.Id,
	)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	if result.RowsAffected() == 0 {
		return repoErrors.ErrorNotFound
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

	result, err := s.db.Exec(
		ctx,
		query,
		status,
		id,
	)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	if result.RowsAffected() == 0 {
		return repoErrors.ErrorNotFound
	}

	return nil
}

func (s *Storage) SaveBuilding(
	ctx context.Context,
	building *model.BuildingModel,
) error {
	query := `INSERT INTO buildings (id, name, description, status, img_url) VALUES ($1, $2, $3, $4, $5);`

	_, err := s.db.Exec(
		ctx,
		query,
		building.Id,
		building.Name,
		building.Description,
		true,
		"/",
	)
	if err != nil {
		return err
	}

	return nil
}
