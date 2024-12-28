package passhandler

import (
	"net/http"
	passService "rip/internal/service/pass"
)

func CompletePassHandler(pService *passService.PassService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		id := r.PathValue("id")
		moderatorToken := ""

		if err := pService.CompletePass(
			r.Context(),
			moderatorToken,
			id,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
