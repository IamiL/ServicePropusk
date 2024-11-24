package buildService

import (
	"fmt"
	model "rip/domain"
	buildRepositoryPostgres "rip/repository/builds/postgres"
	"strconv"
	"strings"
)

type BuildService struct {
	buildRepository     *buildRepositoryPostgres.Storage
	buildImagesHostname string
}

func New(
	buildRep *buildRepositoryPostgres.Storage,
	buildImagesHostname string,
) *BuildService {
	return &BuildService{
		buildRepository:     buildRep,
		buildImagesHostname: buildImagesHostname,
	}
}

func (s *BuildService) GetBuilds(
	serviceName string,
) string {
	services, err := s.buildRepository.Buildings()
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	var servicesHtml string

	if serviceName != "" {
		for _, service := range services {
			if strings.Contains(service.Name, serviceName) {
				serviceIdStr := strconv.Itoa(int(service.Id))
				serviceHtml := `<li><a href="/buildings/` + serviceIdStr + `" class="service-card">` +
					`<img src="` + s.buildImagesHostname + service.ImgUrl + `">` +
					`<h2>` + service.Name + `</h2>
					<form action="add_to_pass/` + serviceIdStr + `" method="post">
					<button type="submit">Добавить в пропуск</button>
					</form>
					</a></li>` + "\n"

				servicesHtml += serviceHtml
			}
		}

		return servicesHtml
	}

	for _, service := range services {
		serviceIdStr := strconv.Itoa(int(service.Id))
		serviceHtml := `<li><a href="/buildings/` + serviceIdStr + `" class="service-card">` +
			`<img src="` + s.buildImagesHostname + service.ImgUrl + `">` +
			`<h2>` + service.Name + `</h2>
			<form action="add_to_pass/` + serviceIdStr + `" method="post">
			<button type="submit">Добавить в пропуск</button>
			</form>
			</a></li>` + "\n"

		servicesHtml += serviceHtml
	}

	return servicesHtml
}

func (s *BuildService) GetBuild(id int64) (model.BuildingModel, error) {
	return s.buildRepository.Build(id)
}

func (s *BuildService) GetBuildImagesHostname() *string {
	return &s.buildImagesHostname
}
