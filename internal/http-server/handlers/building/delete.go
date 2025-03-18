package buildinghandler

import (
	"errors"
	"log/slog"
	"net/http"
	bizErrors "rip/internal/pkg/errors/biz"
	http_api "rip/internal/pkg/http-api"
	buildService "rip/internal/service/building"

	"github.com/google/uuid"
)

// DeleteBuildingHandler godoc
// @Summary Удалить здание
// @Description Удаляет здание из системы
// @Tags buildings
// @Accept json
// @Produce json
// @Param id path string true "ID здания (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings/{id} [delete]
// @Security ApiKeyAuth
func DeleteBuildingHandler(
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

		if err := buildingsService.DeleteBuilding(
			r.Context(),
			accessToken,
			buildingID,
		); err != nil {
			var status int
			switch {
			case errors.Is(err, bizErrors.ErrorNoPermission):
				status = http.StatusForbidden
			case errors.Is(err, bizErrors.ErrorAuthToken):
				status = http.StatusUnauthorized
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
