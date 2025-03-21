package userHandler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	http_api "service-propusk-backend/internal/pkg/http-api"
	"service-propusk-backend/internal/pkg/logger/sl"
	userService "service-propusk-backend/internal/service/user"
)

// EditUserRequest представляет запрос на редактирование пользователя
// @Description Запрос на обновление данных пользователя
type EditUserRequest struct {
	Password string `json:"password" example:"newpassword123"`   // Новый пароль пользователя
	Login    string `json:"login" example:"newuser@example.com"` // Новый email пользователя
}

// EditUserHandler godoc
// @Summary Обновить данные пользователя
// @Description Обновляет информацию о текущем пользователе
// @Tags users
// @Accept json
// @Produce json
// @Param user body EditUserRequest true "Обновленные данные пользователя"
// @Success 200 "OK"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [put]
// @Security ApiKeyAuth
func EditUserHandler(
	log *slog.Logger, uService *userService.UserService,
) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("access_token")
		if err != nil {
			log.Debug("Error getting token", "error", err)
			http_api.HandleError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		accessToken := token.Value

		var req EditUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info("Error decoding body", sl.Err(err))
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Invalid request body",
			)
			return
		}

		if err := uService.Edit(
			r.Context(),
			accessToken,
			req.Login,
			req.Password,
		); err != nil {
			http_api.HandleError(w, http.StatusInternalServerError, err.Error())
		}

	}
}
