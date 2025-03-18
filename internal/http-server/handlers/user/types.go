package userHandler

// ErrorResponse представляет структуру ответа с ошибкой
// @Description Структура ответа при возникновении ошибки
type ErrorResponse struct {
	Message *string `json:"message,omitempty" example:"Описание ошибки"` // Текст ошибки
}
