package buildinghandler

import (
	"errors"
	"log/slog"
	"net/http"
	bizErrors "service-propusk-backend/internal/pkg/errors/biz"

	http_api "rip/internal/pkg/http-api"
	pService "rip/internal/service/pass"

	"github.com/google/uuid"
)

// AddToPassHandler godoc
// @Summary Добавить здание в пропуск
// @Description Добавляет здание в текущий пропуск пользователя
// @Tags buildings
// @Accept json
// @Produce json
// @Param id path string true "ID здания (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings/{id}/draft [post]
// @Security ApiKeyAuth
func AddToPassHandler(
	log *slog.Logger,
	passService *pService.PassService,
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

		if err := passService.AddBuildingToPass(
			r.Context(),
			accessToken,
			buildingID,
		); err != nil {
			var status int
			switch {
			case errors.Is(err, bizErrors.ErrorAuthToken):
				status = http.StatusUnauthorized
			case errors.Is(err, bizErrors.ErrorBuildingNotFound):
				status = http.StatusNotFound
			case errors.Is(err, bizErrors.ErrorBuildingAlreadyAdded):
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
