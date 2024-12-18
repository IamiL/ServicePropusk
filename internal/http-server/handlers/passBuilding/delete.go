package passBuilding

import "net/http"

func DeletePassBuilding() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		passId := r.PathValue("pass_id")

		buildingId := r.PathValue("building_id")

	}
}
