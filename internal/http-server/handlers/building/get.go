package buildinghandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	model "rip/internal/domain"
	buildService "rip/internal/service/build"
	passService "rip/internal/service/pass"
)

type BuildingsResp struct {
	Buildings      *[]model.BuildingModel
	PassID         string `json:"pass_id"`
	PassItemsCount int    `json:"pass_items_count"`
}

func BuildingsHandler(
	buildingsService *buildService.BuildingService,
	passService *passService.PassService,
) func(
	w http.ResponseWriter, r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("BuildingsHandler")
		passID, err := passService.GetPassID(r.Context(), "0")
		if err != nil {
			fmt.Println(err.Error())
		}

		PassItemsCount, err := passService.GetPassItemsCount(r.Context(), "")
		if err != nil {
			fmt.Println(err.Error())
		}

		var buildings *[]model.BuildingModel

		params := r.URL.Query()

		if params.Get("buildName") != "" {
			decodedValue := params.Get("buildName")
			fmt.Println("decoded value:")
			fmt.Println(decodedValue)
			buildings, err = buildingsService.FindBuildings(
				r.Context(),
				decodedValue,
			)
			if err != nil {
				fmt.Println(err.Error())
			}

		} else {
			buildings, err = buildingsService.GetAllBuildings(r.Context())
			if err != nil {
				fmt.Println(err.Error())
			}

		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(
			BuildingsResp{
				PassID:         passID,
				Buildings:      buildings,
				PassItemsCount: PassItemsCount,
			},
		); err != nil {
			http.Error(
				w,
				fmt.Sprintf("error building the response, %v", err),
				http.StatusInternalServerError,
			)
			return
		}

	}
}

type BuildingResp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImgUrl      string `json:"imgUrl"`
}

func BuildingHandler(
	buildingsService *buildService.BuildingService,
) func(
	w http.ResponseWriter, r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("BuildingsHandler2")
		id := r.PathValue("id")

		building, err := buildingsService.GetBuilding(r.Context(), id)
		if err != nil {
			fmt.Println(err.Error())
		}

		if err := json.NewEncoder(w).Encode(
			BuildingResp{
				building.Id,
				building.Name,
				building.Description,
				building.ImgUrl,
			},
		); err != nil {
			http.Error(
				w,
				fmt.Sprintf("error building the response, %v", err),
				http.StatusInternalServerError,
			)
			return
		}
	}
}
