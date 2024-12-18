package handler_mux_v1

import (
	"net/http"
	passService "rip/internal/service/pass"
)

func RejectPassHandler(pService passService.PassService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		id := r.PathValue("id")
		moderatorToken := ""

		if err := pService.RejectPass(
			r.Context(),
			moderatorToken,
			id,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
