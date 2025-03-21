package userHandler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	authService "service-propusk-backend/internal/service/auth"
)

// Credentials представляет учетные данные пользователя
// @Description Учетные данные для входа в систему
type Credentials struct {
	Login    string `json:"login" example:"user@example.com" binding:"required,email"` // Email пользователя
	Password string `json:"password" example:"password123" binding:"required"`         // Пароль пользователя
}

// SigninHandler godoc
// @Summary Войти в систему
// @Description Аутентифицирует пользователя и возвращает токен доступа
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body Credentials true "Учетные данные пользователя"
// @Success 200 "OK"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/login [post]
func SigninHandler(
	log *slog.Logger, authService *authService.AuthService, https bool,
) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sessionToken, err := authService.Auth(
			r.Context(),
			creds.Login,
			creds.Password,
		)
		if err != nil {
			log.Info("error while signing in: " + err.Error())
			w.WriteHeader(http.StatusUnauthorized)
		}

		cookie := &http.Cookie{}

		if https {
			cookie = &http.Cookie{
				Name:     "access_token",
				Value:    sessionToken,
				MaxAge:   30000,
				Path:     "/",
				HttpOnly: true,
				Secure:   true, // Оставьте false для локальной отладки без HTTPS
				SameSite: http.SameSiteLaxMode,
			}
		} else {
			cookie = &http.Cookie{
				Name:     "access_token",
				Value:    sessionToken,
				MaxAge:   30000,
				Path:     "/",
				HttpOnly: true,
				Secure:   false, // Оставьте false для локальной отладки без HTTPS
				SameSite: http.SameSiteLaxMode,
			}
		}

		http.SetCookie(w, cookie)
	}
}
