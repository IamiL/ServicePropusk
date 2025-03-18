package passBuildingHandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	bizErrors "rip/internal/pkg/errors/biz"
	http_api "rip/internal/pkg/http-api"
	"rip/internal/pkg/logger/sl"
	passBuildingService "rip/internal/service/passBuilding"

	"github.com/google/uuid"
)

// EditPassBuildingRequest представляет запрос на редактирование связи пропуска со зданием
// @Description Запрос на обновление информации о связи пропуска со зданием
type EditPassBuildingRequest struct {
	Comment string `json:"comment" example:"Комментарий к зданию в пропуске"` // Комментарий к зданию в пропуске
}

// PutPassBuilding godoc
// @Summary Обновить связь пропуска со зданием
// @Description Обновляет информацию о связи пропуска со зданием (комментарий)
// @Tags pass-buildings
// @Accept json
// @Produce json
// @Param passId path string true "ID пропуска (UUID)"
// @Param buildingId path string true "ID здания (UUID)"
// @Param passBuilding body EditPassBuildingRequest true "Обновленные данные связи"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /passes/{passId}/building/{buildingId}/main [put]
// @Security ApiKeyAuth
func PutPassBuilding(
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

		var req EditPassBuildingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info("Error decoding body", sl.Err(err))
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Invalid request body",
			)
			return
		}

		if err := passBuildingService.Edit(
			r.Context(),
			accessToken,
			passID,
			buildingID,
			req.Comment,
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
