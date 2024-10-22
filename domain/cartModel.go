package model

type CartModel struct {
	Items []ItemModel
	Cost  int
}

type ItemModel struct {
	ServiceID int64
	Quantity  int
}
