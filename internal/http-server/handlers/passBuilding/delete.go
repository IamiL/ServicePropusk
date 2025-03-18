package passBuildingHandler

import (
	"errors"
	"log/slog"
	"net/http"
	bizErrors "rip/internal/pkg/errors/biz"
	http_api "rip/internal/pkg/http-api"
	passBuildingService "rip/internal/service/passBuilding"

	"github.com/google/uuid"
)

// DeletePassBuilding godoc
// @Summary Удалить здание из пропуска
// @Description Удаляет здание из пропуска
// @Tags pass-buildings
// @Accept json
// @Produce json
// @Param passId path string true "ID пропуска (UUID)"
// @Param buildingId path string true "ID здания (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /passes/{passId}/building/{buildingId} [delete]
// @Security ApiKeyAuth
func DeletePassBuilding(
	log *slog.Logger,
	passBuildingService *passBuildingService.PassBuildingService,
) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("access_token")
		if err != nil {
			log.Debug("Error getting token", "error", err)
			http_api.HandleError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		accessToken := token.Value

		passID := r.PathValue("passId")
		if err := uuid.Validate(passID); err != nil {
			http_api.HandleError(w, http.StatusBadRequest, "Invalid pass ID")
			return
		}

		buildingID := r.PathValue("buildingId")
		if err := uuid.Validate(buildingID); err != nil {
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Invalid building ID",
			)
			return
		}

		if err := passBuildingService.Delete(
			r.Context(),
			accessToken,
			passID,
			buildingID,
		); err != nil {
			var status int
			switch {
			case errors.Is(err, bizErrors.ErrorAuthToken):
				status = http.StatusUnauthorized
			case errors.Is(err, bizErrors.ErrorNoPermission):
				status = http.StatusForbidden
			case errors.Is(err, bizErrors.ErrorInvalidPass) || errors.Is(
				err,
				bizErrors.ErrorInvalidPassBuilding,
			) || errors.Is(err, bizErrors.ErrorPassIsNotDraft):
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
