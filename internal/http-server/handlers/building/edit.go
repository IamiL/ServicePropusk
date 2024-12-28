package buildinghandler

import (
	"encoding/json"
	"net/http"
	buildService "rip/internal/service/build"
)

type EditBuildingReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func EditBuildingHandler(
	buildingsService *buildService.BuildingService,
) func(
	w http.ResponseWriter, r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var req EditBuildingReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if err := buildingsService.EditBuilding(
			r.Context(),
			id,
			req.Name,
			req.Description,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
