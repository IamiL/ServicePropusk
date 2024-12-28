package userHandler

import (
	"encoding/json"
	"net/http"
	userService "rip/internal/service/user"
)

type RegistrationRequest struct {
	Password string `json:"password"`
	Login    string `json:"login"`
}

func RegistrationHandler(uService *userService.UserService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegistrationRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := uService.NewUser(
			r.Context(),
			req.Login,
			req.Password,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
