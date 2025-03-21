package passhandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	bizErrors "service-propusk-backend/internal/pkg/errors/biz"
	http_api "service-propusk-backend/internal/pkg/http-api"
	"service-propusk-backend/internal/pkg/logger/sl"
	passService "service-propusk-backend/internal/service/pass"

	"github.com/google/uuid"
)

// ToFormHandler godoc
// @Summary Отправить пропуск на рассмотрение
// @Description Меняет статус пропуска на "submitted" для рассмотрения модератором
// @Tags passes
// @Accept json
// @Produce json
// @Param id path string true "ID пропуска (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /passes/{id}/submit [put]
// @Security ApiKeyAuth
func ToFormHandler(
	log *slog.Logger, pService *passService.PassService,
) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(
		w http.ResponseWriter, r *http.Request,
	) {
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

		if err := pService.ToForm(
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
				bizErrors.ErrorStatusNotDraft,
			) || errors.Is(err, bizErrors.ErrorCannotBeFormed):
				status = http.StatusBadRequest

			default:
				status = http.StatusInternalServerError
			}
			http_api.HandleError(w, status, err.Error())
			return
		}

		pass, _ := pService.Pass(r.Context(), accessToken, passID, true)

		passResp := PassResp{
			ID:          passID,
			VisitorName: pass.VisitorName,
			DateVisit:   pass.DateVisit,
		}

		if len(pass.Items) != 0 {
			passItemsArr := make([]PassItem, 0, len(pass.Items))

			for _, passItem := range pass.Items {
				passItemsArr = append(
					passItemsArr, PassItem{
						Building{
							passItem.Building.Id,
							passItem.Building.Name,
							passItem.Building.Description,
							passItem.Building.ImgUrl,
						},
						passItem.Comment,
						passItem.WasVisited,
					},
				)
			}

			passResp.Items = &passItemsArr
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(
			passResp,
		); err != nil {
			log.Info("Error encoding body: ", sl.Err(err))
		}
	}
}
