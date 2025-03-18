package userHandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	bizErrors "rip/internal/pkg/errors/biz"
	http_api "rip/internal/pkg/http-api"
	"rip/internal/pkg/logger/sl"
	userService "rip/internal/service/user"
)

// RegistrationRequest представляет запрос на регистрацию
// @Description Запрос на регистрацию нового пользователя
type RegistrationRequest struct {
	Login    string `json:"login" example:"user@example.com" binding:"required,email"` // Email пользователя
	Password string `json:"password" example:"password123" binding:"required,min=6"`   // Пароль пользователя (минимум 6 символов)
}

// RegistrationHandler godoc
// @Summary Зарегистрировать нового пользователя
// @Description Создает нового пользователя в системе
// @Tags users
// @Accept json
// @Produce json
// @Param user body RegistrationRequest true "Данные нового пользователя"
// @Success 201 "Created"
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func RegistrationHandler(
	log *slog.Logger, uService *userService.UserService,
) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegistrationRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http_api.HandleError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := uService.NewUser(
			r.Context(),
			req.Login,
			req.Password,
		); err != nil {
			if errors.Is(err, bizErrors.ErrorShortPassword) || errors.Is(
				err,
				bizErrors.ErrorUserAlreadyExists,
			) {
				http_api.HandleError(w, http.StatusBadRequest, err.Error())
				return
			}

			log.Error("RegistrationHandler error", sl.Err(err))
			http_api.HandleError(w, http.StatusInternalServerError, err.Error())
		}
	}
}
