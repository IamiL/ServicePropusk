package buildService

import (
	"context"
	"fmt"
	model "rip/internal/domain"
	postgresBuilds "rip/internal/repository/postgres/builds"
	"strings"
)

type BuildingService struct {
	bProvider           BuildingProvider
	buildImagesHostname string
}

type BuildingProvider interface {
	Building(ctx context.Context, id string) (model.BuildingModel, error)
	Buildings(ctx context.Context) (
		[]model.BuildingModel,
		error,
	)
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

func (s *BuildingService) GetAllBuildingsHTML(
	ctx context.Context,
) string {
	buildings, err := s.bProvider.Buildings(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	var buildingsHtml string

	for _, service := range buildings {
		serviceHtml := `<li><a href="/buildings/` + service.Id + `" class="service-card">` +
			`<img src="` + s.buildImagesHostname + service.ImgUrl + `">` +
			`<h2>` + service.Name + `</h2>
			<form action="add_to_pass/` + service.Id + `" method="post">
			<button type="submit">Добавить в пропуск</button>
			</form>
			</a></li>` + "\n"

		buildingsHtml += serviceHtml
	}

	return buildingsHtml
}

func (s *BuildingService) FindBuildings(
	ctx context.Context,
	buildingName string,
) string {
	if buildingName == "" {
		return s.GetAllBuildingsHTML(ctx)
	}

	buildings, err := s.bProvider.Buildings(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	var buildingsHtml string

	for _, service := range buildings {
		if strings.Contains(service.Name, buildingName) {
			buildingHtml := `<li><a href="/buildings/` + service.Id + `" class="service-card">` +
				`<img src="` + s.buildImagesHostname + service.ImgUrl + `">` +
				`<h2>` + service.Name + `</h2>
					<form action="add_to_pass/` + service.Id + `" method="post">
					<button type="submit">Добавить в пропуск</button>
					</form>
					</a></li>` + "\n"

			buildingsHtml += buildingHtml
		}
	}

	return buildingsHtml
}

func (s *BuildingService) GetBuilding(
	ctx context.Context,
	id string,
) (model.BuildingModel, error) {
	return s.bProvider.Building(ctx, id)
}

func (s *BuildingService) GetBuildImagesHostname() *string {
	return &s.buildImagesHostname
}
