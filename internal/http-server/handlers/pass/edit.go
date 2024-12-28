package passhandler

import (
	"encoding/json"
	"net/http"
	passService "rip/internal/service/pass"
	"time"
)

type EditPassRequest struct {
	Visitor   string    `json:"visitor"`
	DateVisit time.Time `json:"date_visit"`
}

func EditPassHandler(pService *passService.PassService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var req EditPassRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if err := pService.EditPass(
			r.Context(),
			id,
			req.Visitor,
			req.DateVisit,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	}
}
