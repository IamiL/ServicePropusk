package handler_mux_v1

import (
	"net/http"
	passService "rip/internal/service/pass"
)

func ToFormHandler(pService passService.PassService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(
		w http.ResponseWriter, r *http.Request,
	) {
		id := r.PathValue("id")

		if err := pService.ToForm(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
