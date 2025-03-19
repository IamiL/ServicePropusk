package buildinghandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	model "rip/internal/domain"
	bizErrors "rip/internal/pkg/errors/biz"
	http_api "rip/internal/pkg/http-api"
	"rip/internal/pkg/logger/sl"
	buildService "rip/internal/service/building"
	passService "rip/internal/service/pass"

	"github.com/google/uuid"
)

// BuildingsResp представляет ответ со списком зданий
// @Description Ответ, содержащий список зданий и информацию о текущем пропуске
type BuildingsResp struct {
	Buildings      *[]model.BuildingModel `json:"buildings"`                  // Список зданий
	PassID         *string                `json:"pass_id,omitempty"`          // ID текущего пропуска
	PassItemsCount *int                   `json:"pass_items_count,omitempty"` // Количество зданий в пропуске
}

// BuildingsHandler godoc
// @Summary Получить список зданий
// @Description Возвращает список всех зданий с возможностью фильтрации по названию
// @Tags buildings
// @Accept json
// @Produce json
// @Param buildName query string false "Название здания для фильтрации"
// @Success 200 {object} BuildingsResp
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings [get]
// @Security ApiKeyAuth
func BuildingsHandler(
	log *slog.Logger,
	buildingsService *buildService.BuildingService,
	passService *passService.PassService,
) func(
	w http.ResponseWriter, r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("get BuildingsHandler")

		var passIDResp *string
		var PassItemsCountResp *int

		passID, err := passService.GetPassID(r.Context(), "")
		if err != nil {
			log.Info("Error getting pass ID", sl.Err(err))
		} else {
			if len(passID) != 0 {
				passIDResp = &passID
			}
		}

		passItemsCount, err := passService.GetPassItemsCount(
			r.Context(),
			"",
		)
		if err != nil {
			log.Info("Error getting pass items count", sl.Err(err))
		} else {
			PassItemsCountResp = &passItemsCount
		}

		var buildings *[]model.BuildingModel
		params := r.URL.Query()

		if buildName := params.Get("buildName"); buildName != "" {
			log.Debug("decoded value", "value", buildName)
			buildings, err = buildingsService.FindBuildings(
				r.Context(),
				buildName,
			)
			if err != nil {
				if errors.Is(err, bizErrors.ErrorBuildingsNotFound) {
					http_api.HandleError(
						w,
						http.StatusNotFound,
						"Buildings not found",
					)
					return
				}
				http_api.HandleError(
					w,
					http.StatusInternalServerError,
					"Error finding buildings",
				)
				return
			}
		} else {
			buildings, err = buildingsService.GetAllBuildings(r.Context())
			if err != nil {
				http_api.HandleError(
					w,
					http.StatusInternalServerError,
					"Error retrieving all buildings",
				)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(
			BuildingsResp{
				Buildings:      buildings,
				PassID:         passIDResp,
				PassItemsCount: PassItemsCountResp,
			},
		); err != nil {
			http_api.HandleError(
				w,
				http.StatusInternalServerError,
				"Error encoding response",
			)
		}
	}
}

// BuildingResp представляет ответ с информацией о здании
// @Description Ответ, содержащий информацию о конкретном здании
type BuildingResp struct {
	ID          string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"` // ID здания
	Name        string `json:"name" example:"Офисное здание"`                     // Название здания
	Description string `json:"description" example:"Многоэтажное офисное здание"` // Описание здания
	ImgUrl      string `json:"imgUrl" example:"http://example.com/image.jpg"`     // URL изображения здания
}

// BuildingHandler godoc
// @Summary Получить здание по ID
// @Description Возвращает информацию о конкретном здании
// @Tags buildings
// @Accept json
// @Produce json
// @Param id path string true "ID здания (UUID)"
// @Success 200 {object} BuildingResp
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings/{id} [get]
// @Security ApiKeyAuth
func BuildingHandler(
	buildingsService *buildService.BuildingService,
) func(
	w http.ResponseWriter, r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		buildingID := r.PathValue("id")
		if err := uuid.Validate(buildingID); err != nil {
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Invalid building ID",
			)
			return
		}

		building, err := buildingsService.GetBuilding(r.Context(), buildingID)
		if err != nil {
			if errors.Is(err, bizErrors.ErrorBuildingNotFound) {
				http_api.HandleError(
					w,
					http.StatusNotFound,
					"Building not found",
				)
				return
			}
			http_api.HandleError(
				w,
				http.StatusInternalServerError,
				"Error finding buildings",
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(
			BuildingResp{
				building.Id,
				building.Name,
				building.Description,
				building.ImgUrl,
			},
		); err != nil {
			http_api.HandleError(
				w,
				http.StatusInternalServerError,
				"Error encoding response",
			)
		}
	}
}
