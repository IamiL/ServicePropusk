package postgresPasses

import (
	"context"
	"errors"
	"fmt"
	model "rip/internal/domain"
	repoErrors "rip/internal/pkg/errors/repo"
	passService "rip/internal/service/pass"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (*Storage, error) {
	return &Storage{db: pool}, nil
}

func (s *Storage) PassShort(ctx context.Context, id string) (
	*model.PassModel,
	error,
) {
	const op = "repository.passes.postgres.PassShort"

	query := `SELECT status, creator FROM passes WHERE id = $1`

	pass := model.PassModel{}

	err := s.db.QueryRow(ctx, query, id).Scan(
		&pass.Status,
		&pass.CreatorID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PassModel{}, repoErrors.ErrorNotFound
		}
		return &model.PassModel{}, fmt.Errorf(
			"%s: execute statement: %w",
			op,
			err,
		)
	}

	return &pass, nil
}

func (s *Storage) Pass(ctx context.Context, id string) (
	*model.PassModel,
	error,
) {
	const op = "repository.passes.postgres.Pass"

	query := `SELECT creator, visitor, visit_date, status FROM passes WHERE id = $1`

	pass := model.PassModel{}

	err := s.db.QueryRow(ctx, query, id).Scan(
		&pass.CreatorID,
		&pass.VisitorName,
		&pass.DateVisit,
		&pass.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("error errnorows")
			return &model.PassModel{}, repoErrors.ErrorNotFound
		}
		return &model.PassModel{}, fmt.Errorf(
			"%s: execute statement: %w",
			op,
			err,
		)
	}

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

	pass.ID = id

	pass.Items = make(model.PassItems, 0, 5)

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
	query := `UPDATE passes SET status = 1, formation_date = $1 WHERE id = $2;`

	_, err := s.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return nil
}

func (s *Storage) NewPass(
	ctx context.Context,
	id string,
	uid string,
	status int,
	visitor string,
	visitDate time.Time,
) error {
	query := `INSERT INTO passes (id, creator, created_at, visitor, visit_date, status) VALUES ($1, $2, $3, $4, $5, $6);`

	if _, err := s.db.Exec(
		ctx,
		query,
		id,
		uid,
		time.Now(),
		visitor,
		visitDate,
		status,
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

func (s *Storage) PassesForUser(
	ctx context.Context, uid string, statusFilter *int,
	beginDateFilter *time.Time,
	endDateFilter *time.Time,
) (*[]passService.PassModel, error) {
	const op = "repository.passes.postgres.Passes"

	fmt.Println("\n=== PassesForUser ===")
	fmt.Println("Input parameters:")
	fmt.Printf("uid: %s\n", uid)
	fmt.Printf("statusFilter: %v\n", statusFilter)
	fmt.Printf("beginDateFilter: %v\n", beginDateFilter)
	fmt.Printf("endDateFilter: %v\n", endDateFilter)

	query := `SELECT p.id, u.login, m.login, p.visitor, p.visit_date, p.status, p.formed_at 
			FROM passes p 
			LEFT JOIN users m ON p.moderator = m.id 
			JOIN users u ON p.creator = u.id 
			WHERE p.creator = $1`

	args := []interface{}{uid}
	argNum := 2

	fmt.Println("\nBuilding query:")
	fmt.Printf("Base query: %s\n", query)

	if statusFilter != nil {
		query += fmt.Sprintf(" AND p.status = $%d", argNum)
		args = append(args, *statusFilter)
		argNum++
		fmt.Printf("Added status filter: %s\n", query)
	} else {
		query += " AND NOT p.status = 0 AND NOT p.status = 4"
		fmt.Printf("Added default status filter: %s\n", query)
	}

	if beginDateFilter != nil && !beginDateFilter.IsZero() {
		query += fmt.Sprintf(" AND p.visit_date >= $%d", argNum)
		args = append(args, *beginDateFilter)
		argNum++
		fmt.Printf("Added begin date filter: %s\n", query)
	}

	if endDateFilter != nil && !endDateFilter.IsZero() {
		query += fmt.Sprintf(" AND p.visit_date <= $%d", argNum)
		args = append(args, *endDateFilter)
		fmt.Printf("Added end date filter: %s\n", query)
	}

	fmt.Println("\nFinal query and args:")
	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Args: %v\n", args)
	fmt.Println("===================\n")

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	services := make([]passService.PassModel, 0)
	for rows.Next() {
		c := passService.PassModel{}
		var moderator pgtype.Text

		err := rows.Scan(
			&c.ID,
			&c.Creator.Login,
			&moderator,
			&c.VisitorName,
			&c.DateVisit,
			&c.Status,
			&c.FormedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: execute statement: %w", op, err)
		}

		if moderator.Valid {
			c.Moderator = &passService.User{Login: moderator.String}
		}

		services = append(services, c)
	}

	return &services, nil
}

func (s *Storage) Passes(
	ctx context.Context, statusFilter *int,
	beginDateFilter *time.Time,
	endDateFilter *time.Time,
) (*[]passService.PassModel, error) {
	const op = "repository.passes.postgres.Passes"

	fmt.Println("\n=== Passes ===")
	fmt.Println("Input parameters:")
	fmt.Printf("statusFilter: %v\n", statusFilter)
	fmt.Printf("beginDateFilter: %v\n", beginDateFilter)
	fmt.Printf("endDateFilter: %v\n", endDateFilter)

	query := `SELECT p.id, u.login, m.login, p.visitor, p.visit_date, p.status, p.formed_at 
			FROM passes p 
			LEFT JOIN users m ON p.moderator = m.id 
			JOIN users u ON p.creator = u.id 
			WHERE 1=1`

	args := make([]interface{}, 0)
	argNum := 1

	fmt.Println("\nBuilding query:")
	fmt.Printf("Base query: %s\n", query)

	if statusFilter != nil {
		query += fmt.Sprintf(" AND p.status = $%d", argNum)
		args = append(args, *statusFilter)
		argNum++
		fmt.Printf("Added status filter: %s\n", query)
	} else {
		query += " AND NOT p.status = 0 AND NOT p.status = 4"
		fmt.Printf("Added default status filter: %s\n", query)
	}

	if beginDateFilter != nil && !beginDateFilter.IsZero() {
		query += fmt.Sprintf(" AND p.visit_date >= $%d", argNum)
		args = append(args, *beginDateFilter)
		argNum++
		fmt.Printf("Added begin date filter: %s\n", query)
	}

	if endDateFilter != nil && !endDateFilter.IsZero() {
		query += fmt.Sprintf(" AND p.visit_date <= $%d", argNum)
		args = append(args, *endDateFilter)
		fmt.Printf("Added end date filter: %s\n", query)
	}

	fmt.Println("\nFinal query and args:")
	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Args: %v\n", args)
	fmt.Println("===================\n")

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	services := make([]passService.PassModel, 0)
	for rows.Next() {
		c := passService.PassModel{}
		var moderator pgtype.Text

		err := rows.Scan(
			&c.ID,
			&c.Creator.Login,
			&moderator,
			&c.VisitorName,
			&c.DateVisit,
			&c.Status,
			&c.FormedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: execute statement: %w", op, err)
		}

		if moderator.Valid {
			c.Moderator = &passService.User{Login: moderator.String}
		}

		services = append(services, c)
	}

	return &services, nil
}

func (s *Storage) EditPass(
	ctx context.Context,
	id string,
	visitor string,
	visitDate time.Time,
) error {
	query := `UPDATE passes SET visitor = $1, visit_date = $2 WHERE id = $3;`

	_, err := s.db.Exec(ctx, query, visitor, visitDate, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) EditPassStatusByModerator(
	ctx context.Context,
	id string,
	status int,
	time time.Time,
	moderatorId string,
) error {
	query := `UPDATE passes SET status = $1, moderator = $2, completed_at = $3 WHERE id = $4;`

	_, err := s.db.Exec(ctx, query, status, moderatorId, time, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) EditPassStatusByUser(
	ctx context.Context,
	id string,
	status int,
	time time.Time,
) error {
	query := `UPDATE passes SET status = $1, formed_at = $2 WHERE id = $3;`

	_, err := s.db.Exec(ctx, query, status, time, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) EditWasVisitedForPass(ctx context.Context, id string) error {
	query := `UPDATE buildings_passes SET was_used = true WHERE pass = $1;`

	_, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) EditPassBuildingComment(
	ctx context.Context,
	passID string,
	buildingID string,
	newComment string,
) error {
	query := `UPDATE buildings_passes SET comment = $1 WHERE pass = $2 AND building = $3;`

	result, err := s.db.Exec(ctx, query, newComment, passID, buildingID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repoErrors.ErrorNotFound
	}

	return nil
}

func (s *Storage) DeleteBuildingFromPass(
	ctx context.Context,
	passID string,
	buildingID string,
) error {
	query := `DELETE from buildings_passes WHERE building = $1 AND pass = $2`

	result, err := s.db.Exec(ctx, query, buildingID, passID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return repoErrors.ErrorNotFound
	}
	return nil
}
