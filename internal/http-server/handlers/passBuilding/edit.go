package passBuildingHandler

import (
	"encoding/json"
	"net/http"
	passBuildingService "rip/internal/service/passBuilding"
)

type EditPassBuildingRequest struct {
	Comment string `json:"comment"`
}

func PutPassBuilding(passBuildingService *passBuildingService.PassBuildingService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		passID := r.PathValue("passId")

		buildingID := r.PathValue("buildingId")

		var req EditPassBuildingRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if err := passBuildingService.Edit(
			r.Context(),
			passID,
			buildingID,
			req.Comment,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
