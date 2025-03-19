package buildinghandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	bizErrors "rip/internal/pkg/errors/biz"
	http_api "rip/internal/pkg/http-api"
	"rip/internal/pkg/logger/sl"
	buildService "rip/internal/service/building"
)

// NewBuildingReq представляет запрос на создание нового здания
// @Description Запрос на создание нового здания
type NewBuildingReq struct {
	Name        string `json:"name" example:"Офисное здание" binding:"required"`  // Название здания
	Description string `json:"description" example:"Многоэтажное офисное здание"` // Описание здания
}

// NewBuildingHandler godoc
// @Summary Создать новое здание
// @Description Создает новое здание в системе
// @Tags buildings
// @Accept json
// @Produce json
// @Param building body NewBuildingReq true "Данные нового здания"
// @Success 201 "Created"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings [post]
// @Security ApiKeyAuth
func NewBuildingHandler(
	log *slog.Logger,
	buildingsService *buildService.BuildingService,
) func(
	w http.ResponseWriter, r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req NewBuildingReq

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info("Error decoding body", sl.Err(err))
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Invalid request body",
			)
			return
		}

		if err := buildingsService.AddBuilding(
			r.Context(),
			"",
			req.Name,
			req.Description,
		); err != nil {
			var status int
			switch {
			case errors.Is(err, bizErrors.ErrorAuthToken):
				status = http.StatusUnauthorized
			case errors.Is(err, bizErrors.ErrorNoPermission):
				status = http.StatusForbidden
			case errors.Is(err, bizErrors.ErrorInvalidBuilding):
				status = http.StatusBadRequest
			default:
				status = http.StatusInternalServerError
			}
			http_api.HandleError(w, status, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
