package passBuildingHandler

import (
	"net/http"
	passBuildingService "rip/internal/service/passBuilding"
)

func DeletePassBuilding(passBuildingService *passBuildingService.PassBuildingService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		passID := r.PathValue("passId")

		buildingID := r.PathValue("buildingId")

		if err := passBuildingService.Delete(
			r.Context(),
			passID,
			buildingID,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
