package cartRepository

import (
	"errors"
	model "rip/domain"
	"sync"
)

type Repository struct {
	Storage map[int64]model.CartModel
	sync.RWMutex
}

func New() *Repository {
	carts := make(map[int64]model.CartModel, 5)

	carts[0] = model.CartModel{
		Items: []model.ItemModel{{1, 2}, {4, 1}, {3, 5}},
		Cost:  100,
	}
	carts[1] = model.CartModel{
		Items: []model.ItemModel{{3, 2}, {2, 6}, {4, 1}},
		Cost:  300,
	}

	return &Repository{Storage: carts}
}

func (r *Repository) Cart(id int64) (model.CartModel, error) {
	cart, exists := r.Storage[id]
	if !exists {
		return cart, errors.New("cart not found")
	}

	return cart, nil
}
