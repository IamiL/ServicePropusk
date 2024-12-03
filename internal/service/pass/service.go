package passService

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"rip/internal/consts"
	model "rip/internal/domain"
	postgresPasses "rip/internal/repository/postgres/passes"
	"time"
)

type PassService struct {
	passProvider        PassProvider
	passSaver           PassSaver
	passDeleter         PassDeleter
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
}

type PassSaver interface {
	AddToPass(
		ctx context.Context,
		recordID string,
		id string,
		buildingID string,
	) error
	NewDraftPass(
		ctx context.Context,
		id string,
		uid string,
		visitor string,
		visitDate time.Time,
	) error
}

type PassDeleter interface {
	Delete(ctx context.Context, id string) error
}

func New(
	passReporitory *postgresPasses.Storage,
	buildImagesHostname string,
) *PassService {
	return &PassService{
		passProvider:        passReporitory,
		passSaver:           passReporitory,
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

func (p *PassService) GetPassHTML(
	ctx context.Context,
	id string,
) (*string, error) {
	fmt.Println("getPassHTML service start")

	pass, err := p.passProvider.Pass(ctx, id)
	if err != nil {
		return nil, err
	}

	fmt.Println("getPassHTML service, builds in pass: ", len(pass.Items))
	fmt.Println("getPassHTML service end")

	return pass.GetHMTL(&p.buildImagesHostname), nil
}

func (p *PassService) AddToPass(
	ctx context.Context,
	token string,
	build string,
) error {
	userID := consts.UserID

	passID, err := p.passProvider.DraftPassIDByCreator(ctx, userID)
	if err != nil {
		passID = uuid.NewString()

		err = p.passSaver.NewDraftPass(ctx, passID, userID, "", time.Now())
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	recordID := uuid.NewString()

	return p.passSaver.AddToPass(ctx, recordID, passID, build)
}

func (p *PassService) Delete(ctx context.Context, id string) error {
	return p.passDeleter.Delete(ctx, id)
}

func (p *PassService) GetPassItemsCount(ctx context.Context, token string) (
	int,
	error,
) {
	userID := consts.UserID

	return p.passProvider.ItemsCount(ctx, userID)
}
