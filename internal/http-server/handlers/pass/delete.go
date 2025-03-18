package passhandler

import (
	"errors"
	"log/slog"
	"net/http"
	bizErrors "rip/internal/pkg/errors/biz"
	http_api "rip/internal/pkg/http-api"
	passService "rip/internal/service/pass"

	"github.com/google/uuid"
)

// DeletePassHandler godoc
// @Summary Удалить пропуск
// @Description Удаляет пропуск из системы
// @Tags passes
// @Accept json
// @Produce json
// @Param id path string true "ID пропуска (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /passes/{id} [delete]
// @Security ApiKeyAuth
func DeletePassHandler(
	log *slog.Logger, pService *passService.PassService,
) func(
	http.ResponseWriter, *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("access_token")
		if err != nil {
			log.Debug("Error getting token", "error", err)
			http_api.HandleError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		accessToken := token.Value

		passID := r.PathValue("id")
		if err := uuid.Validate(passID); err != nil {
			http_api.HandleError(w, http.StatusBadRequest, "Invalid pass ID")
			return
		}

		if err := pService.Delete(
			r.Context(),
			accessToken,
			passID,
		); err != nil {
			var status int
			switch {
			case errors.Is(err, bizErrors.ErrorAuthToken):
				status = http.StatusUnauthorized
			case errors.Is(err, bizErrors.ErrorNoPermission):
				status = http.StatusForbidden
			case errors.Is(err, bizErrors.ErrorInvalidPass) || errors.Is(
				err,
				bizErrors.ErrorCannotBeDeleted,
			):
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
