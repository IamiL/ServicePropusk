package passService

import (
	"fmt"
	passRepositoryPostgres "rip/repository/passes/postgres"
	buildService "rip/service/build"
	"strconv"
	"time"
)

type PassService struct {
	passReporitory      *passRepositoryPostgres.Storage
	buildService        *buildService.BuildService
	buildImagesHostname string
}

func New(
	passReporitory *passRepositoryPostgres.Storage,
	buildService *buildService.BuildService,
	buildImagesHostname string,
) *PassService {
	return &PassService{
		passReporitory:      passReporitory,
		buildService:        buildService,
		buildImagesHostname: buildImagesHostname,
	}
}

func (s *PassService) GetPassID(token string) (string, error) {
	id, err := s.passReporitory.PassID(0)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(int(id)), nil
}

func (p *PassService) GetPassHTML(
	id int64,
) (*string, error) {
	fmt.Println("getPassHTML service start")
	pass, err := p.passReporitory.Pass(id)
	if err != nil {
		return nil, err
	}

	fmt.Println("getPassHTML service, builds in pass: ", len(pass.Items))
	fmt.Println("getPassHTML service end")

	return pass.GetHMTL(&p.buildImagesHostname), nil
}

func (p *PassService) AddToPass(uid int64, build int64) error {
	passID, err := p.passReporitory.FindDraftPassByCreator(uid)
	if err != nil {
		passID, err = p.passReporitory.NewDraftPass(uid, "", time.Now())
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	return p.passReporitory.AddToPass(passID, build)
}

func (p *PassService) Delete(id int64) error {
	return p.passReporitory.Delete(id)
}
