package buildinghandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	buildService "rip/internal/service/build"
)

type NewBuildingReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewBuildingHandler(
	buildingsService *buildService.BuildingService,
) func(
	w http.ResponseWriter, r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("BuildingsHandler3")
		var req NewBuildingReq

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := buildingsService.AddBuilding(
			r.Context(),
			req.Name,
			req.Description,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	}
}
