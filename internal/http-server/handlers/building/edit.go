package buildinghandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	bizErrors "service-propusk-backend/internal/pkg/errors/biz"
	http_api "service-propusk-backend/internal/pkg/http-api"
	"service-propusk-backend/internal/pkg/logger/sl"
	buildService "service-propusk-backend/internal/service/building"

	"github.com/google/uuid"
)

// EditBuildingRequest представляет запрос на редактирование здания
// @Description Запрос на обновление информации о здании
type EditBuildingRequest struct {
	Name        string `json:"name" example:"Офисное здание"`                     // Название здания
	Description string `json:"description" example:"Многоэтажное офисное здание"` // Описание здания
}

// EditBuildingHandler godoc
// @Summary Обновить здание
// @Description Обновляет информацию о существующем здании
// @Tags buildings
// @Accept json
// @Produce json
// @Param id path string true "ID здания (UUID)"
// @Param building body EditBuildingRequest true "Обновленные данные здания"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings/{id} [put]
// @Security ApiKeyAuth
func EditBuildingHandler(
	log *slog.Logger,
	buildingsService *buildService.BuildingService,
) func(
	w http.ResponseWriter, r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("access_token")
		if err != nil {
			log.Debug("Error getting token", "error", err)
			http_api.HandleError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		accessToken := token.Value

		buildingID := r.PathValue("id")
		if err := uuid.Validate(buildingID); err != nil {
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Invalid building ID",
			)
			return
		}

		var req EditBuildingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info("Error decoding body: ", sl.Err(err))
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Invalid request body",
			)
			return
		}

		if err := buildingsService.EditBuilding(
			r.Context(),
			accessToken,
			buildingID,
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

		building, _ := buildingsService.GetBuilding(r.Context(), buildingID)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(
			BuildingResp{
				building.Id,
				building.Name,
				building.Description,
				building.ImgUrl,
			},
		); err != nil {
			log.Info("Error encoding body: ", sl.Err(err))
		}
	}
}
