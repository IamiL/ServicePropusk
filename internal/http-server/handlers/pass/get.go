package passhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	bizErrors "rip/internal/pkg/errors/biz"
	http_api "rip/internal/pkg/http-api"
	passService "rip/internal/service/pass"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// PassShortResp представляет краткую информацию о пропуске
// @Description Краткая информация о пропуске для списка
type PassShortResp struct {
	User      User       `json:"user"`                                              // Информация о создателе пропуска
	Moderator *User      `json:"moderator,omitempty"`                               // Информация о модераторе (если есть)
	ID        string     `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"` // ID пропуска
	Status    int        `json:"status" example:"1"`                                // Статус пропуска
	FormedAt  *time.Time `json:"formed_at,omitempty"`                               // Дата формирования
}

// User представляет информацию о пользователе
// @Description Информация о пользователе
type User struct {
	Login string `json:"login" example:"user@example.com"` // Логин пользователя
}

// PassesHandler godoc
// @Summary Получить список пропусков
// @Description Возвращает список пропусков с возможностью фильтрации по статусу и датам
// @Tags passes
// @Accept json
// @Produce json
// @Param status query int false "Статус пропуска для фильтрации"
// @Param start_date query string false "Начальная дата для фильтрации (формат: DD.MM.YYYY)"
// @Param end_date query string false "Конечная дата для фильтрации (формат: DD.MM.YYYY)"
// @Success 200 {array} PassShortResp
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /passes [get]
// @Security ApiKeyAuth
func PassesHandler(
	log *slog.Logger, pService *passService.PassService,
) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		var statusFilter *int = nil

		if params.Get("status") != "" {
			status, err := strconv.Atoi(params.Get("status"))
			if err != nil {

			} else {
				statusFilter = &status
			}
		}

		var beginDateFilter *time.Time = nil

		if params.Get("start_date") != "" {
			beginDateFilterStr := params.Get("start_date")
			if utf8.RuneCountInString(beginDateFilterStr) < 10 {
				http_api.HandleError(
					w,
					http.StatusBadRequest,
					"Invalid date format",
				)
				return
			}

			var timeTemp time.Time
			var err error

			// Try YYYY-MM-DD format first
			timeTemp, err = time.Parse("2006-01-02", beginDateFilterStr)
			if err != nil {
				// If failed, try DD.MM.YYYY format
				day, err := strconv.Atoi(beginDateFilterStr[0:2])
				if err != nil {
					http_api.HandleError(
						w,
						http.StatusBadRequest,
						"Invalid day in date",
					)
					return
				}

				month, err := strconv.Atoi(beginDateFilterStr[3:5])
				if err != nil {
					http_api.HandleError(
						w,
						http.StatusBadRequest,
						"Invalid month in date",
					)
					return
				}

				year, err := strconv.Atoi(beginDateFilterStr[6:10])
				if err != nil {
					http_api.HandleError(
						w,
						http.StatusBadRequest,
						"Invalid year in date",
					)
					return
				}

				timeTemp = time.Date(
					year,
					time.Month(month),
					day,
					0,
					0,
					0,
					0,
					time.UTC,
				)
			}

			beginDateFilter = &timeTemp
		}

		var endDateFilter *time.Time = nil

		if params.Get("end_date") != "" {
			endDateFilterStr := params.Get("end_date")
			if utf8.RuneCountInString(endDateFilterStr) < 10 {
				http_api.HandleError(
					w,
					http.StatusBadRequest,
					"Invalid date format",
				)
				return
			}

			var timeTemp time.Time
			var err error

			// Try YYYY-MM-DD format first
			timeTemp, err = time.Parse("2006-01-02", endDateFilterStr)
			if err != nil {
				// If failed, try DD.MM.YYYY format
				day, err := strconv.Atoi(endDateFilterStr[0:2])
				if err != nil {
					http_api.HandleError(
						w,
						http.StatusBadRequest,
						"Invalid day in date",
					)
					return
				}

				month, err := strconv.Atoi(endDateFilterStr[3:5])
				if err != nil {
					http_api.HandleError(
						w,
						http.StatusBadRequest,
						"Invalid month in date",
					)
					return
				}

				year, err := strconv.Atoi(endDateFilterStr[6:10])
				if err != nil {
					http_api.HandleError(
						w,
						http.StatusBadRequest,
						"Invalid year in date",
					)
					return
				}

				timeTemp = time.Date(
					year,
					time.Month(month),
					day,
					23,
					59,
					59,
					999999999,
					time.UTC,
				)
			} else {
				// If YYYY-MM-DD format was successful, set time to end of day
				timeTemp = timeTemp.Add(24*time.Hour - time.Nanosecond)
			}

			endDateFilter = &timeTemp
		}

		passes, err := pService.Passes(
			r.Context(),
			"",
			statusFilter,
			beginDateFilter,
			endDateFilter,
		)
		if err != nil {
			var status int
			switch {
			case errors.Is(err, bizErrors.ErrorAuthToken):
				status = http.StatusUnauthorized
			case errors.Is(err, bizErrors.ErrorNoPermission):
				status = http.StatusForbidden
			case errors.Is(err, bizErrors.ErrorPassesNotFound):
				status = http.StatusNotFound
			default:
				status = http.StatusInternalServerError
			}
			http_api.HandleError(w, status, err.Error())
			return
		}

		resp := make([]PassShortResp, 0, len(*passes))

		for _, p := range *passes {
			var moderator *User
			if p.Moderator != nil {
				moderator = &User{p.Moderator.Login}
			}
			resp = append(
				resp,
				PassShortResp{
					User{p.Creator.Login},
					moderator,
					p.ID,
					p.Status,
					p.FormedAt,
				},
			)
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(
			resp,
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

// PassResp представляет полную информацию о пропуске
// @Description Полная информация о пропуске
type PassResp struct {
	ID          string      `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`  // ID пропуска
	Items       *[]PassItem `json:"items,omitempty"`                                    // Список зданий в пропуске
	VisitorName string      `json:"visitorName,omitempty" example:"Иван Иванов"`        // Имя посетителя
	DateVisit   *time.Time  `json:"dateVisit,omitempty" example:"2024-03-10T12:00:00Z"` // Дата посещения
}

// PassItem представляет элемент пропуска
// @Description Элемент пропуска (здание с комментарием)
type PassItem struct {
	Building Building `json:"building"`                               // Информация о здании
	Comment  string   `json:"comment" example:"Комментарий к зданию"` // Комментарий к зданию
}

// Building представляет информацию о здании в пропуске
// @Description Информация о здании в пропуске
type Building struct {
	Id          string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"` // ID здания
	Name        string `json:"name" example:"Офисное здание"`                     // Название здания
	Description string `json:"description" example:"Многоэтажное офисное здание"` // Описание здания
	ImgUrl      string `json:"imgUrl" example:"http://example.com/image.jpg"`     // URL изображения здания
}

// PassHandler godoc
// @Summary Получить пропуск по ID
// @Description Возвращает полную информацию о конкретном пропуске
// @Tags passes
// @Accept json
// @Produce json
// @Param id path string true "ID пропуска (UUID)"
// @Success 200 {object} PassResp
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /passes/{id} [get]
// @Security ApiKeyAuth
func PassHandler(log *slog.Logger, pService *passService.PassService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		passID := r.PathValue("id")
		if err := uuid.Validate(passID); err != nil {
			http_api.HandleError(w, http.StatusBadRequest, "Invalid pass ID")
			return
		}

		pass, err := pService.Pass(r.Context(), "", passID)
		if err != nil {
			if errors.Is(err, bizErrors.ErrorPassNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

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
					},
				)
			}

			passResp.Items = &passItemsArr
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(
			passResp,
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
