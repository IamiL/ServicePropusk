package userHandler

import (
	"encoding/json"
	"net/http"
	userService "rip/internal/service/user"
)

type EditUserRequest struct {
	Password string `json:"password"`
	Login    string `json:"login"`
}

func EditUserHandler(uService *userService.UserService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var req EditUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if err := uService.Edit(
			r.Context(),
			token.Value,
			req.Login,
			req.Password,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	}
}
