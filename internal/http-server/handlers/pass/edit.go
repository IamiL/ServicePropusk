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
	"time"

	"github.com/google/uuid"
)

// EditPassRequest представляет запрос на редактирование пропуска
// @Description Запрос на обновление информации о пропуске
type EditPassRequest struct {
	Visitor   string    `json:"visitor" example:"Иван Иванов"`             // Имя посетителя
	DateVisit time.Time `json:"date_visit" example:"2024-03-10T12:00:00Z"` // Дата посещения
}

// EditPassHandler godoc
// @Summary Обновить пропуск
// @Description Обновляет информацию о существующем пропуске
// @Tags passes
// @Accept json
// @Produce json
// @Param id path string true "ID пропуска (UUID)"
// @Param pass body EditPassRequest true "Обновленные данные пропуска"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /passes/{id} [put]
// @Security ApiKeyAuth
func EditPassHandler(
	log *slog.Logger, pService *passService.PassService,
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

		passID := r.PathValue("id")
		if err := uuid.Validate(passID); err != nil {
			http_api.HandleError(w, http.StatusBadRequest, "Invalid pass ID")
			return
		}

		var req EditPassRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info("Error decoding body", sl.Err(err))
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Invalid request body",
			)
			return
		}

		if err = pService.EditPass(
			r.Context(),
			accessToken,
			passID,
			req.Visitor,
			req.DateVisit,
		); err != nil {
			var status int
			switch {
			case errors.Is(err, bizErrors.ErrorAuthToken):
				status = http.StatusUnauthorized
			case errors.Is(err, bizErrors.ErrorNoPermission):
				status = http.StatusForbidden
			case errors.Is(err, bizErrors.ErrorInvalidPass) || errors.Is(
				err,
				bizErrors.ErrorCannotBeEditing,
			):
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
