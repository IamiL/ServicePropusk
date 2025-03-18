package buildingService

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	model "rip/internal/domain"
	bizErrors "rip/internal/pkg/errors/biz"
	repoErrors "rip/internal/pkg/errors/repo"
	"rip/internal/pkg/logger/sl"
	postgresBuilds "rip/internal/repository/postgres/buildings"

	"github.com/google/uuid"
)

type BuildingService struct {
	log                 *slog.Logger
	bProvider           BuildingProvider
	bSaver              BuildingSaver
	bEditor             BuildingEditor
	s3bDeleter          s3BuildingDeleter
	s3bSaver            s3BuildingSaver
	authService         AuthService
	buildImagesHostname string
	buildImagesBucket   string
}

type BuildingProvider interface {
	Building(ctx context.Context, id string) (model.BuildingModel, error)
	AllBuildings(ctx context.Context) (
		[]model.BuildingModel,
		error,
	)
	FindBuildings(ctx context.Context, name string) (
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
	EditBuildingImgUrl(ctx context.Context, id string, url string) error
}

type s3BuildingDeleter interface {
	DeleteBuildingPreview(ctx context.Context, id string) error
}

type s3BuildingSaver interface {
	SaveBuildingPreview(
		ctx context.Context,
		id string,
		object []byte,
	) error
}

type AuthService interface {
	Claims(token string) (string, bool, error)
}

func New(
	log *slog.Logger,
	buildingRep *postgresBuilds.Storage,
	bSaver BuildingSaver,
	bEditor BuildingEditor,
	s3bDeleter s3BuildingDeleter,
	s3bSaver s3BuildingSaver,
	authService AuthService,
	buildImagesHostname string,
) *BuildingService {
	return &BuildingService{
		log:                 log,
		bProvider:           buildingRep,
		bSaver:              bSaver,
		bEditor:             bEditor,
		s3bDeleter:          s3bDeleter,
		s3bSaver:            s3bSaver,
		authService:         authService,
		buildImagesHostname: buildImagesHostname,
		buildImagesBucket:   "services", // Используем фиксированное имя бакета
	}
}

func (s *BuildingService) GetAllBuildings(
	ctx context.Context,
) (*[]model.BuildingModel, error) {
	buildings, err := s.bProvider.AllBuildings(ctx)
	if err != nil {
		s.log.Error("error get all buildings: ", sl.Err(err))
		return nil, bizErrors.ErrorInternalServer
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

	buildings, err := s.bProvider.FindBuildings(ctx, buildingName)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return nil, bizErrors.ErrorBuildingsNotFound
		}
		s.log.Error("FindBuildings error: ", sl.Err(err))

		return nil, bizErrors.ErrorInternalServer
	}

	return &buildings, nil
}

func (s *BuildingService) GetBuilding(
	ctx context.Context,
	id string,
) (model.BuildingModel, error) {
	building, err := s.bProvider.Building(ctx, id)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return model.BuildingModel{}, bizErrors.ErrorBuildingNotFound
		}

		s.log.Error("error get building service: ", sl.Err(err))
		return model.BuildingModel{}, bizErrors.ErrorInternalServer
	}

	return building, nil
}

func (s *BuildingService) AddBuilding(
	ctx context.Context,
	accessToken string,
	name string,
	description string,
) error {
	_, isAdmin, err := s.authService.Claims(accessToken)
	if err != nil {
		return bizErrors.ErrorAuthToken
	}

	if !isAdmin {
		s.log.Info("I don't have enough rights to edit the case.")
		return bizErrors.ErrorNoPermission
	}

	building := model.BuildingModel{
		Id:          uuid.NewString(),
		Name:        name,
		Description: description,
	}

	err = s.bSaver.SaveBuilding(ctx, &building)
	if err != nil {
		s.log.Error("error new building: ", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	return nil
}

func (s *BuildingService) EditBuilding(
	ctx context.Context,
	accessToken string,
	buildingID string,
	name string,
	description string,
) error {
	_, isAdmin, err := s.authService.Claims(accessToken)
	if err != nil {
		return bizErrors.ErrorAuthToken
	}

	if !isAdmin {
		s.log.Info("I don't have enough rights to edit the case.")
		return bizErrors.ErrorNoPermission
	}

	err = s.bEditor.EditBuildingInfo(
		ctx,
		&model.BuildingModel{
			Id:          buildingID,
			Name:        name,
			Description: description,
		},
	)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			s.log.Info("Invalid building.")
			return bizErrors.ErrorInvalidBuilding
		}

		s.log.Error("error editing building: ", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	return nil
}

func (s *BuildingService) DeleteBuilding(
	ctx context.Context,
	accessToken string,
	buildingID string,
) error {
	_, isAdmin, err := s.authService.Claims(accessToken)
	if err != nil {
		return err
	}

	if !isAdmin {
		s.log.Info("I don't have enough rights to delete the case.")
		return bizErrors.ErrorNoPermission
	}

	_, err = s.bProvider.Building(ctx, buildingID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			s.log.Info("Invalid building.")
			return bizErrors.ErrorInvalidBuilding
		}
	}

	err = s.bEditor.EditBuildingStatus(
		ctx,
		buildingID,
		false,
	)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			s.log.Info("Invalid building.")
			return bizErrors.ErrorInvalidBuilding
		}

		s.log.Error("error editing building status: ", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	err = s.s3bDeleter.DeleteBuildingPreview(ctx, buildingID)
	if err != nil {
		s.log.Error("error deleting building preview: ", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	return nil
}

func (s *BuildingService) EditBuildingPreview(
	ctx context.Context,
	accessToken string,
	id string,
	photo []byte,
) error {
	// Проверяем права доступа
	_, isAdmin, err := s.authService.Claims(accessToken)
	if err != nil {
		s.log.Error("failed to check auth claims", sl.Err(err))
		return bizErrors.ErrorAuthToken
	}

	if !isAdmin {
		s.log.Info("user doesn't have permission to edit building preview")
		return bizErrors.ErrorNoPermission
	}

	// Проверяем существование здания
	_, err = s.bProvider.Building(ctx, id)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			s.log.Info("building not found", "id", id)
			return bizErrors.ErrorBuildingNotFound
		}
		s.log.Error("failed to get building", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	s.log.Info("deleting old building preview", "id", id)
	if err := s.s3bDeleter.DeleteBuildingPreview(ctx, id); err != nil {
		s.log.Warn("failed to delete old preview, continuing", sl.Err(err))
	}

	s.log.Info("saving new building preview", "id", id)
	if err := s.s3bSaver.SaveBuildingPreview(ctx, id, photo); err != nil {
		s.log.Error("failed to save preview", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	// Обновляем URL изображения в базе данных
	imgUrl := fmt.Sprintf("/%s/%s.png", s.buildImagesBucket, id)
	s.log.Info("updating building image URL", "id", id, "url", imgUrl)

	if err := s.bEditor.EditBuildingImgUrl(ctx, id, imgUrl); err != nil {
		s.log.Error("failed to update image URL", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	return nil
}

func (s *BuildingService) CheckEditAccess(
	ctx context.Context,
	accessToken string,
) error {
	_, isAdmin, err := s.authService.Claims(accessToken)
	if err != nil {
		return bizErrors.ErrorAuthToken
	}

	if !isAdmin {
		return bizErrors.ErrorNoPermission
	}

	return nil
}

func (s *BuildingService) GetBuildImagesHostname() *string {
	return &s.buildImagesHostname
}
