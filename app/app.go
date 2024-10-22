package app

import (
	ginApi "rip/controller/http/v1/gin"
	cartRepository "rip/repository/carts/inmemory"
	serviceRepository "rip/repository/services/inmemory"
)

type App struct {
}

func New() *App {
	return &App{}
}

func (*App) MustRun() {
	servicesRepository := serviceRepository.New()

	cartsRepository := cartRepository.New()

	ginApi.StartServer(servicesRepository, cartsRepository)
}
