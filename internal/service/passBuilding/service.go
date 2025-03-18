package passBuildingService

import (
	"context"
	"errors"
	"log/slog"
	"rip/internal/consts"
	model "rip/internal/domain"
	bizErrors "rip/internal/pkg/errors/biz"
	repoErrors "rip/internal/pkg/errors/repo"
	"rip/internal/pkg/logger/sl"
)

type PassBuildingService struct {
	log          *slog.Logger
	passProvider PassProvider
	passEditor   PassEditor
	authService  AuthService
}

type PassProvider interface {
	PassShort(ctx context.Context, id string) (*model.PassModel, error)
}

type PassEditor interface {
	EditPassBuildingComment(
		ctx context.Context,
		passID string,
		buildingID string,
		newComment string,
	) error

	DeleteBuildingFromPass(
		ctx context.Context,
		passID string,
		buildingID string,
	) error
}

type AuthService interface {
	Claims(token string) (string, bool, error)
}

func New(
	log *slog.Logger, passEditor PassEditor,
	passProvider PassProvider,
	authService AuthService,
) *PassBuildingService {
	return &PassBuildingService{
		log:          log,
		passProvider: passProvider,
		passEditor:   passEditor,
		authService:  authService,
	}
}

func (s *PassBuildingService) Edit(
	ctx context.Context,
	accessToken string,
	passID string,
	buildingID string,
	newComment string,
) error {
	uid, _, err := s.authService.Claims(accessToken)
	if err != nil {
		return err
	}

	pass, err := s.passProvider.PassShort(ctx, passID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPass
		}

		s.log.Error("error edit passbuilding get pass: ", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	if pass.CreatorID != uid {
		return bizErrors.ErrorInvalidPass
	}

	if pass.Status != consts.StatusDraft {
		return bizErrors.ErrorPassIsNotDraft
	}

	if err := s.passEditor.EditPassBuildingComment(
		ctx,
		passID,
		buildingID,
		newComment,
	); err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPassBuilding
		}

		s.log.Error("error edit passBuilding comment: ", sl.Err(err))

		return bizErrors.ErrorInternalServer
	}

	return nil
}

func (s *PassBuildingService) Delete(
	ctx context.Context,
	accessToken string,
	passID string,
	buildingID string,
) error {
	uid, _, err := s.authService.Claims(accessToken)
	if err != nil {
		return err
	}

	pass, err := s.passProvider.PassShort(ctx, passID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPass
		}

		s.log.Error("error complete pass: ", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	if pass.CreatorID != uid {
		return bizErrors.ErrorInvalidPass
	}

	if pass.Status != consts.StatusDraft {
		return bizErrors.ErrorPassIsNotDraft
	}

	if err := s.passEditor.DeleteBuildingFromPass(
		ctx,
		passID,
		buildingID,
	); err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPassBuilding
		}

		s.log.Error("error delete passBuilding: ", sl.Err(err))

		return bizErrors.ErrorInternalServer
	}

	return nil
}
