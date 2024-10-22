package serviceRepository

import (
	"errors"
	model "rip/domain"
	"sync"
)

type Repository struct {
	Storage []model.ServiceModel
	sync.RWMutex
}

func New() *Repository {
	services := make([]model.ServiceModel, 5)

	services[0] = model.ServiceModel{
		0,
		"Главный корпус (ГУК)",
		"Оформление пропуска в главный учебный корпус по адресу 2-я Бауманская улица, 5",
		`/services/0.png`,
		150,
	}
	services[1] = model.ServiceModel{
		1,
		"Учебно-лабораторный корпус",
		"Оформление пропуска в учебно-лабораторный корпус по адресу 2-я Бауманская улица, 5",
		`/services/1.png`,
		180,
	}
	services[2] = model.ServiceModel{
		2,
		"Корпус Э",
		`Оформление пропуска в корпус "энерго" по адресу 2-я Бауманская улица, 5`,
		`/services/2.png`,
		90,
	}
	services[3] = model.ServiceModel{
		3,
		"Корпус СМ",
		`Оформление пропуска в корпус "специальное машиностроение" по адресу 2-я Бауманская улица, 5`,
		`/services/3.png`,
		60,
	}
	services[4] = model.ServiceModel{
		4,
		"Корпус Т",
		`Оформление пропуска в корпус "т" по адресу 2-я Бауманская улица, 5`,
		`/services/4.png`,
		95,
	}

	return &Repository{Storage: services}
}

func (r *Repository) Services() []model.ServiceModel {
	return r.Storage
}

func (r *Repository) Service(id int64) (model.ServiceModel, error) {
	for _, service := range r.Storage {
		if service.Id == id {
			return service, nil
		}
	}

	return model.ServiceModel{}, errors.New("service not found")
}
