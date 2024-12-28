package passService

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"rip/internal/consts"
	model "rip/internal/domain"
	"time"
)

const (
	StatusDraft     = 0
	StatusFormed    = 1
	StatusReject    = 2
	StatusCompleted = 3
	StatusDeleted   = 4
)

type PassService struct {
	passProvider        PassProvider
	passSaver           PassSaver
	passDeleter         PassDeleter
	passEditor          PassEditor
	buildImagesHostname string
}

type PassProvider interface {
	ID(ctx context.Context, uid string) (string, error)
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
		visitDate time.Time,
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

func New(
	passProvider PassProvider,
	passSaver PassSaver,
	passDeleter PassDeleter,
	passEditor PassEditor,
	buildImagesHostname string,
) *PassService {
	return &PassService{
		passProvider:        passProvider,
		passSaver:           passSaver,
		passDeleter:         passDeleter,
		passEditor:          passEditor,
		buildImagesHostname: buildImagesHostname,
	}
}

func (p *PassService) GetPassID(ctx context.Context, token string) (
	string,
	error,
) {
	userID := consts.UserID

	id, err := p.passProvider.ID(ctx, userID)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (p *PassService) Pass(
	ctx context.Context,
	id string,
) (*model.PassModel, error) {

	pass, err := p.passProvider.Pass(ctx, id)
	if err != nil {
		return nil, err
	}

	fmt.Println("getPass service, builds in pass: ", len(pass.Items))

	return pass, nil
}

func (p *PassService) DeleteBuildingFromPass(
	ctx context.Context,
	token string,
	buildingId string,
	passId string,
) error {
	err := p.passEditor.DeleteBuildingFromPass(ctx, buildingId, passId)
	if err != nil {
		return err
	}

	return nil
}

func (p *PassService) AddBuildingToPass(
	ctx context.Context,
	token string,
	build string,
) error {
	userID := consts.UserID

	passID, err := p.passProvider.DraftPassIDByCreator(ctx, userID)
	if err != nil {
		passID = uuid.NewString()

		err = p.passSaver.NewPass(
			ctx,
			passID,
			userID,
			StatusDraft,
			"",
			time.Now(),
		)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	recordID := uuid.NewString()

	return p.passSaver.AddToPass(ctx, recordID, passID, build)
}

func (p *PassService) Delete(
	ctx context.Context,
	token string,
	id string,
) error {
	if err := p.passEditor.EditPassStatusByUser(
		ctx,
		id,
		StatusDeleted,
		time.Now(),
	); err != nil {
		return err
	}

	return nil
}

func (p *PassService) GetPassItemsCount(ctx context.Context, token string) (
	int,
	error,
) {
	userID := consts.UserID

	return p.passProvider.ItemsCount(ctx, userID)
}

func (p *PassService) Passes(
	ctx context.Context,
	statusFilter *int,
	beginDateFilter *time.Time,
	endDateFilter *time.Time,
) (*[]PassModel, error) {
	passes, err := p.passProvider.Passes(
		ctx, statusFilter, beginDateFilter,
		endDateFilter,
	)
	if err != nil {
		return nil, err
	}

	return passes, nil
}

type PassModel struct {
	User        User
	ID          string
	VisitorName string
	DateVisit   time.Time
	Status      int
}

type User struct {
	Id    string
	Login string
}

func (p *PassService) EditPass(
	ctx context.Context,
	id string,
	visitor string,
	dateVisit time.Time,
) error {
	if err := p.passEditor.EditPass(ctx, id, visitor, dateVisit); err != nil {
		fmt.Println("err: ", err.Error())
		return err
	}

	return nil
}

func (p *PassService) ToForm(
	ctx context.Context,
	id string,
) error {
	pass, err := p.passProvider.Pass(ctx, id)
	if err != nil {
		return err
	}
	if pass.VisitorName == "" {
		return errors.New("no visitor name")
	}

	if err := p.passEditor.EditPassStatusByUser(
		ctx,
		id,
		StatusFormed,
		time.Now(),
	); err != nil {
		return err
	}

	return nil
}

func (p *PassService) RejectPass(
	ctx context.Context,
	token string,
	id string,
) error {
	moderatorId := ""

	time := time.Now()

	if err := p.passEditor.EditPassStatusByModerator(
		ctx, id, StatusReject, time, moderatorId,
	); err != nil {
		return err
	}

	if err := p.passEditor.EditWasVisitedForPass(
		context.Background(),
		id,
	); err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

func (p *PassService) CompletePass(
	ctx context.Context,
	token string,
	id string,
) error {
	moderatorId := ""

	if err := p.passEditor.EditPassStatusByModerator(
		ctx, id, StatusCompleted, time.Now(), moderatorId,
	); err != nil {
		return err
	}

	if err := p.passEditor.EditWasVisitedForPass(
		context.Background(),
		id,
	); err != nil {
		fmt.Println(err.Error())
	}

	return nil
}
