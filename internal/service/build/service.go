package buildService

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"io"
	model "rip/internal/domain"
	postgresBuilds "rip/internal/repository/postgres/builds"
)

type BuildingService struct {
	bProvider           BuildingProvider
	bSaver              BuildingSaver
	bEditor             BuildingEditor
	s3bDeleter          s3BuildingDeleter
	s3bSaver            s3BuildingSaver
	buildImagesHostname string
}

type BuildingProvider interface {
	Building(ctx context.Context, id string) (model.BuildingModel, error)
	Buildings(ctx context.Context) (
		[]model.BuildingModel,
		error,
	)
}

type BuildingSaver interface {
	SaveBuilding(ctx context.Context, building *model.BuildingModel) error
}

type BuildingEditor interface {
	EditBuildingInfo(ctx context.Context, building *model.BuildingModel) error
	EditBuildingStatus(
		ctx context.Context,
		id string,
		status bool,
	) error
}

type s3BuildingDeleter interface {
	DeleteBuildingPreview(ctx context.Context, id string) error
}

type s3BuildingSaver interface {
	SaveBuildingPreview(
		ctx context.Context,
		id string,
		object io.Reader,
	) error
}

func New(
	buildingRep *postgresBuilds.Storage,
	buildImagesHostname string,
) *BuildingService {
	return &BuildingService{
		bProvider:           buildingRep,
		buildImagesHostname: buildImagesHostname,
	}
}

func (s *BuildingService) GetAllBuildings(
	ctx context.Context,
) (*[]model.BuildingModel, error) {
	buildings, err := s.bProvider.Buildings(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &buildings, nil
}

func (s *BuildingService) FindBuildings(
	ctx context.Context,
	buildingName string,
) (*[]model.BuildingModel, error) {
	if buildingName == "" {
		return s.GetAllBuildings(ctx)
	}

	buildings, err := s.bProvider.Buildings(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &buildings, nil
}

func (s *BuildingService) GetBuilding(
	ctx context.Context,
	id string,
) (model.BuildingModel, error) {
	return s.bProvider.Building(ctx, id)
}

func (s *BuildingService) AddBuilding(
	ctx context.Context,
	name string,
	description string,
) error {
	building := model.BuildingModel{
		Id:          uuid.NewString(),
		Name:        name,
		Description: description,
	}

	if err := s.bSaver.SaveBuilding(ctx, &building); err != nil {
		return err
	}

	return nil
}

func (s *BuildingService) EditBuilding(
	ctx context.Context,
	id string,
	name string,
	description string,
) error {
	if err := s.bEditor.EditBuildingInfo(
		ctx,
		&model.BuildingModel{Id: id, Name: name, Description: description},
	); err != nil {
		return err
	}

	return nil
}

func (s *BuildingService) DeleteBuilding(ctx context.Context, id string) error {
	if err := s.bEditor.EditBuildingStatus(ctx, id, false); err != nil {
		return err
	}

	if err := s.s3bDeleter.DeleteBuildingPreview(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *BuildingService) EditBuildingPreview(
	ctx context.Context,
	id string,
	photo io.Reader,
) error {
	if err := s.s3bDeleter.DeleteBuildingPreview(ctx, id); err != nil {
		//return err
	}

	if err := s.s3bSaver.SaveBuildingPreview(ctx, id, photo); err != nil {
		return err
	}

	return nil
}

func (s *BuildingService) GetBuildImagesHostname() *string {
	return &s.buildImagesHostname
}
