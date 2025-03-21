package userHandler

import (
	"log/slog"
	"net/http"
	userService "service-propusk-backend/internal/service/user"
	"time"
)

// LogoutHandler godoc
// @Summary Выйти из системы
// @Description Очищает токен доступа пользователя и завершает сессию
// @Tags users
// @Produce json
// @Success 200 "OK"
// @Failure 500 {object} ErrorResponse
// @Router /users/logout [post]
// @Security ApiKeyAuth
func LogoutHandler(
	log *slog.Logger, uService *userService.UserService,
) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		// immediately clear the token cookie
		http.SetCookie(
			w, &http.Cookie{
				Name:    "access_token",
				Expires: time.Now(),
			},
		)
	}
}
