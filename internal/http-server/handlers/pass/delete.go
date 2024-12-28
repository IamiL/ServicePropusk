package passhandler

import (
	"net/http"
	passService "rip/internal/service/pass"
)

func DeletePassHandler(pService *passService.PassService) func(
	http.ResponseWriter, *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		token := ""

		id := r.PathValue("id")

		if err := pService.Delete(r.Context(), token, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
