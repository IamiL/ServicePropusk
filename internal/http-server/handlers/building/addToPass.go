package handler_mux_v1

import (
	"log/slog"
	"net/http"
	passService "rip/internal/service/pass"
)

func AddToPassHandler(
	log *slog.Logger, passService *passService.PassService,
) func(
	w http.ResponseWriter, r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		//const op = "handlers.building.addToPass.AddToPassHandler"
		//
		//log := log.With(
		//	slog.String("op", op),
		//)

		id := r.PathValue("id")

		//c, err := r.Cookie("session_token")
		//if err != nil {
		//	if err == http.ErrNoCookie {
		//		// If the cookie is not set, return an unauthorized status
		//		w.WriteHeader(http.StatusUnauthorized)
		//		return
		//	}
		//	// For any other type of error, return a bad request status
		//	w.WriteHeader(http.StatusBadRequest)
		//	return
		//}
		sessionToken := ""

		if err := passService.AddBuildingToPass(
			r.Context(),
			sessionToken,
			id,
		); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	}
}
