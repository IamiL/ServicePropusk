package passService

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"rip/internal/consts"
	model "rip/internal/domain"
	bizErrors "rip/internal/pkg/errors/biz"
	repoErrors "rip/internal/pkg/errors/repo"
	"rip/internal/pkg/logger/sl"
	"time"

	"github.com/google/uuid"

	"github.com/skip2/go-qrcode"
)

type PassService struct {
	log                 *slog.Logger
	passProvider        PassProvider
	passSaver           PassSaver
	passDeleter         PassDeleter
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

type PassDeleter interface {
	Delete(ctx context.Context, id string) error
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
	passDeleter PassDeleter,
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
		passDeleter:         passDeleter,
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
	id, err := p.passProvider.ID(ctx, "155ce20e-b039-4851-a775-2bf4d1e38c24")
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			p.log.Info("pass not found, uid: ")
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
) (*model.PassModel, error) {
	pass, err := p.passProvider.Pass(ctx, id)
	if err != nil {
		return nil, err
	}

	if pass.Status == consts.StatusDeleted {
		return nil, bizErrors.ErrorPassNotFound
	}

	return pass, nil
}

func (p *PassService) AddBuildingToPass(
	ctx context.Context,
	accessToken string,
	buildingID string,
) error {
	_, err := p.bProvider.Building(ctx, buildingID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			p.log.Info("Building not found.")
			return bizErrors.ErrorBuildingNotFound
		}

		return bizErrors.ErrorInternalServer
	}

	passID, err := p.passProvider.DraftPassIDByCreator(
		ctx,
		"155ce20e-b039-4851-a775-2bf4d1e38c24",
	)
	if err != nil {
		passID = uuid.NewString()
		var visit_date *time.Time

		err = p.passSaver.NewPass(
			ctx,
			passID,
			"155ce20e-b039-4851-a775-2bf4d1e38c24",
			consts.StatusDraft,
			"",
			visit_date,
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
	pass, err := p.passProvider.PassShort(ctx, passID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPass
		}
	}

	if pass.Status != consts.StatusDraft {
		p.log.Info("status pass to deleted: ", "status: ", pass.Status)
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
	return p.passProvider.ItemsCount(
		ctx,
		"155ce20e-b039-4851-a775-2bf4d1e38c24",
	)
}

func (p *PassService) Passes(
	ctx context.Context,
	accessToken string,
	statusFilter *int,
	beginDateFilter *time.Time,
	endDateFilter *time.Time,
) (*[]PassModel, error) {

	p.log.Info("запрос на пропуска")
	if statusFilter != nil {
		p.log.Info("фильтр", "status", *statusFilter)
	} else {
		p.log.Info("без фильтра по статусу")
	}
	if beginDateFilter != nil {
		p.log.Info("фильтр", "beginDate", *beginDateFilter)
	} else {
		p.log.Info("без фильтра по beginDate")
	}
	if endDateFilter != nil {
		p.log.Info("фильтр", "endDate", *endDateFilter)
	} else {
		p.log.Info("без фильтра по endDate")
	}

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
	pass, err := p.passProvider.PassShort(ctx, passID)
	if err != nil {
		if errors.Is(err, repoErrors.ErrorNotFound) {
			return bizErrors.ErrorInvalidPass
		}
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
		ctx,
		passID,
		consts.StatusCompleted,
		time.Now(),
		"155ce20e-b039-4851-a775-2bf4d1e38c24",
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

	// Generate QR code
	qr, err := qrcode.New(
		fmt.Sprintf(
			"https://172.20.10.5:3000/passes/%s",
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
