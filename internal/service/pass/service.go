package passService

import (
	"context"
	"errors"
	"fmt"
	"github.com/skip2/go-qrcode"
	"log/slog"
	"service-propusk-backend/internal/consts"
	model "service-propusk-backend/internal/domain"
	bizErrors "service-propusk-backend/internal/pkg/errors/biz"
	repoErrors "service-propusk-backend/internal/pkg/errors/repo"
	"service-propusk-backend/internal/pkg/logger/sl"
	"time"

	"github.com/google/uuid"
)

type PassService struct {
	log                 *slog.Logger
	passProvider        PassProvider
	passSaver           PassSaver
	passEditor          PassEditor
	bProvider           BuildingProvider
	authService         AuthService
	buildImagesHostname string
	qrCodeSaver         QRCodeSaver
}

type PassProvider interface {
	ID(ctx context.Context, uid string) (string, error)
	PassShort(ctx context.Context, id string) (*model.PassModel, error)
	Pass(ctx context.Context, id string) (*model.PassModel, error)
	DraftPassIDByCreator(
		ctx context.Context,
		uid string,
	) (string, error)
	ItemsCount(ctx context.Context, uid string) (
		int,
		error,
	)
	Passes(
		ctx context.Context, statusFilter *int,
		beginDateFilter *time.Time,
		endDateFilter *time.Time,
	) (*[]PassModel, error)
	PassesForUser(
		ctx context.Context, uid string, statusFilter *int,
		beginDateFilter *time.Time,
		endDateFilter *time.Time,
	) (*[]PassModel, error)
}

type PassSaver interface {
	AddToPass(
		ctx context.Context,
		recordID string,
		id string,
		buildingID string,
	) error
	NewPass(
		ctx context.Context,
		id string,
		uid string,
		status int,
		visitor string,
		visitDate *time.Time,
	) error
}

type PassEditor interface {
	DeleteBuildingFromPass(
		ctx context.Context,
		buildingID string,
		passId string,
	) error

	EditPass(
		ctx context.Context,
		id string,
		visitor string,
		visitDate time.Time,
	) error

	EditPassStatusByModerator(
		ctx context.Context,
		id string,
		status int,
		time time.Time,
		moderatorId string,
	) error

	EditPassStatusByUser(
		ctx context.Context,
		id string,
		status int,
		time time.Time,
	) error

	EditWasVisitedForPass(
		ctx context.Context,
		id string,
	) error
}

type BuildingProvider interface {
	Building(ctx context.Context, id string) (model.BuildingModel, error)
}

type AuthService interface {
	Claims(token string) (string, bool, error)
}

type QRCodeSaver interface {
	SaveQRCode(ctx context.Context, id string, qrCode []byte) error
}

func New(
	log *slog.Logger,
	passProvider PassProvider,
	passSaver PassSaver,
	passEditor PassEditor,
	buildingProvider BuildingProvider,
	authService AuthService,
	buildImagesHostname string,
	qrCodeSaver QRCodeSaver,
) *PassService {
	return &PassService{
		log:                 log,
		passProvider:        passProvider,
		passSaver:           passSaver,
		passEditor:          passEditor,
		bProvider:           buildingProvider,
		authService:         authService,
		buildImagesHostname: buildImagesHostname,
		qrCodeSaver:         qrCodeSaver,
	}
}

func (p *PassService) GetPassID(ctx context.Context, accessToken string) (
	string,
	error,
) {
	uid, _, err := p.authService.Claims(accessToken)
	if err != nil {
		p.log.Info("Error getting token claims: ", sl.Err(err))
		return "", bizErrors.ErrorAuthToken
	}

	id, err := p.passProvider.ID(ctx, uid)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			p.log.Info("pass not found, uid: ", uid)
			return "", bizErrors.ErrorPassesNotFound
		}

		p.log.Error("Error getting passID: ", sl.Err(err))
		return "", bizErrors.ErrorInternalServer
	}

	return id, nil
}

func (p *PassService) Pass(
	ctx context.Context,
	accessToken string,
	id string,
	protected bool,
) (*model.PassModel, error) {
	uid, isAdmin, err := p.authService.Claims(accessToken)
	if err != nil && protected {
		return nil, err
	}

	pass, err := p.passProvider.Pass(ctx, id)
	if err != nil {
		return nil, err
	}

	if pass.CreatorID != uid && !isAdmin && protected {
		p.log.Info("недостаточно прав для просмотра заявки")
		return nil, bizErrors.ErrorNoPermission
	}

	return pass, nil
}

func (p *PassService) AddBuildingToPass(
	ctx context.Context,
	accessToken string,
	buildingID string,
) error {
	uid, _, err := p.authService.Claims(accessToken)
	if err != nil {
		return err
	}

	_, err = p.bProvider.Building(ctx, buildingID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			p.log.Info("Building not found.")
			return bizErrors.ErrorBuildingNotFound
		}

		return bizErrors.ErrorInternalServer
	}

	passID, err := p.passProvider.DraftPassIDByCreator(ctx, uid)
	if err != nil {
		passID = uuid.NewString()

		err = p.passSaver.NewPass(
			ctx,
			passID,
			uid,
			consts.StatusDraft,
			"",
			nil,
		)
		if err != nil {
			p.log.Error("error: ", sl.Err(err))
			return bizErrors.ErrorInternalServer
		}
	} else {
		pass, err := p.passProvider.Pass(ctx, passID)
		if err != nil {
			p.log.Error("error: ", sl.Err(err))
			return bizErrors.ErrorInternalServer
		}

		for _, item := range pass.Items {
			if item.Building.Id == buildingID {
				return bizErrors.ErrorBuildingAlreadyAdded
			}
		}
	}

	recordID := uuid.NewString()

	return p.passSaver.AddToPass(ctx, recordID, passID, buildingID)
}

func (p *PassService) Delete(
	ctx context.Context,
	accessToken string,
	passID string,
) error {
	uid, _, err := p.authService.Claims(accessToken)
	if err != nil {
		return err
	}

	pass, err := p.passProvider.PassShort(ctx, passID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPass
		}
	}

	if pass.CreatorID != uid {
		return bizErrors.ErrorNoPermission
	}

	if pass.Status != consts.StatusDraft {
		return bizErrors.ErrorCannotBeDeleted
	}

	if err := p.passEditor.EditPassStatusByUser(
		ctx,
		passID,
		consts.StatusDeleted,
		time.Now(),
	); err != nil {
		p.log.Error("error delete pass: ", sl.Err(err))

		return bizErrors.ErrorInternalServer
	}

	return nil
}

func (p *PassService) GetPassItemsCount(
	ctx context.Context,
	accessToken string,
) (
	int,
	error,
) {
	uid, _, err := p.authService.Claims(accessToken)
	if err != nil {
		p.log.Info("Error getting token claims: ", sl.Err(err))
	}

	return p.passProvider.ItemsCount(ctx, uid)
}

func (p *PassService) Passes(
	ctx context.Context,
	accessToken string,
	statusFilter *int,
	beginDateFilter *time.Time,
	endDateFilter *time.Time,
) (*[]PassModel, error) {
	uid, isAdmin, err := p.authService.Claims(accessToken)
	if err != nil {
		return nil, err
	}

	if isAdmin {
		passes, err := p.passProvider.Passes(
			ctx, statusFilter, beginDateFilter,
			endDateFilter,
		)
		if err != nil {
			if errors.Is(err, repoErrors.ErrorNotFound) {
				return nil, bizErrors.ErrorPassesNotFound
			}

			p.log.Error("error get passes service: ", sl.Err(err))

			return nil, bizErrors.ErrorInternalServer
		}

		return passes, nil
	} else {
		passes, err := p.passProvider.PassesForUser(
			ctx, uid, statusFilter, beginDateFilter,
			endDateFilter,
		)
		if err != nil {
			if errors.Is(err, repoErrors.ErrorNotFound) {
				return nil, bizErrors.ErrorPassesNotFound
			}

			p.log.Error("error get passes service: ", sl.Err(err))

			return nil, bizErrors.ErrorInternalServer
		}

		return passes, nil
	}
}

type PassModel struct {
	ID          string
	Creator     User
	Moderator   *User
	VisitorName string
	DateVisit   time.Time
	Status      int
	FormedAt    *time.Time
}

type User struct {
	Login string
}

func (p *PassService) EditPass(
	ctx context.Context,
	accessToken string,
	passID string,
	visitor string,
	dateVisit time.Time,
) error {
	uid, _, err := p.authService.Claims(accessToken)
	if err != nil {
		return err
	}

	pass, err := p.passProvider.PassShort(ctx, passID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPass
		}
	}

	if pass.CreatorID != uid {
		return bizErrors.ErrorNoPermission
	}

	if pass.Status == consts.StatusCompleted {
		return bizErrors.ErrorCannotBeEditing
	}

	if err := p.passEditor.EditPass(
		ctx,
		passID,
		visitor,
		dateVisit,
	); err != nil {
		p.log.Error("error edit pass: ", sl.Err(err))

		return bizErrors.ErrorInternalServer
	}

	return nil
}

func (p *PassService) ToForm(
	ctx context.Context,
	accessToken string,
	passID string,
) error {
	uid, _, err := p.authService.Claims(accessToken)
	if err != nil {
		return err
	}

	pass, err := p.passProvider.Pass(ctx, passID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPass
		}

		p.log.Error("error complete pass: ", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	if pass.Status != consts.StatusDraft {
		return bizErrors.ErrorStatusNotDraft
	}

	if pass.CreatorID != uid {
		return bizErrors.ErrorNoPermission
	}

	if len(pass.Items) == 0 {
		return fmt.Errorf(
			"%w: в пропуске нет корпусов",
			bizErrors.ErrorCannotBeFormed,
		)
	}

	if pass.VisitorName == "" {
		return fmt.Errorf(
			"%w: нет ФИО посетителя",
			bizErrors.ErrorCannotBeFormed,
		)
	}

	if err := p.passEditor.EditPassStatusByUser(
		ctx,
		passID,
		consts.StatusFormed,
		time.Now(),
	); err != nil {
		return err
	}

	// Generate QR code
	qr, err := qrcode.New(
		fmt.Sprintf(
			"https://172.17.17.145:3000/pass/%s/info",
			passID,
		), qrcode.Medium,
	)
	if err != nil {
		p.log.Error("failed to generate QR code", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	// Get PNG bytes
	png, err := qr.PNG(256)
	if err != nil {
		p.log.Error("failed to encode QR code as PNG", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	// Save QR code to MinIO
	if err := p.qrCodeSaver.SaveQRCode(ctx, passID, png); err != nil {
		p.log.Error("failed to save QR code to storage", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	return nil
}

func (p *PassService) RejectPass(
	ctx context.Context,
	accessToken string,
	passID string,
) error {
	uid, isAdmin, err := p.authService.Claims(accessToken)
	if err != nil {
		return err
	}

	if !isAdmin {
		p.log.Info("I don't have enough rights.")
		return bizErrors.ErrorNoPermission
	}

	pass, err := p.passProvider.PassShort(ctx, passID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPass
		}

		p.log.Error("error complete pass: ", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	if pass.Status != consts.StatusFormed {
		return bizErrors.ErrorStatusNotFormed
	}

	if err := p.passEditor.EditPassStatusByModerator(
		ctx, passID, consts.StatusReject, time.Now(), uid,
	); err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPass
		}

		p.log.Error("error complete pass: ", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	return nil
}

func (p *PassService) CompletePass(
	ctx context.Context,
	accessToken string,
	passID string,
) error {
	uid, isAdmin, err := p.authService.Claims(accessToken)
	if err != nil {
		return err
	}

	if !isAdmin {
		p.log.Info("I don't have enough rights.")
		return bizErrors.ErrorNoPermission
	}

	pass, err := p.passProvider.PassShort(ctx, passID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPass
		}

		p.log.Error("error complete pass: ", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	if pass.Status != consts.StatusFormed {
		return bizErrors.ErrorStatusNotFormed
	}

	if err := p.passEditor.EditPassStatusByModerator(
		ctx, passID, consts.StatusCompleted, time.Now(), uid,
	); err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPass
		}

		p.log.Error("error complete pass: ", sl.Err(err))
		return bizErrors.ErrorInternalServer
	}

	if err := p.passEditor.EditWasVisitedForPass(
		context.Background(),
		passID,
	); err != nil {
		p.log.Error("error save was visited for pass: ", sl.Err(err))
	}

	return nil
}
