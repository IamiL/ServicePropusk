package buildinghandler

import (
	"io"
	"log/slog"
	"net/http"
	http_api "rip/internal/pkg/http-api"
	"rip/internal/pkg/logger/sl"
	buildService "rip/internal/service/building"

	"github.com/google/uuid"
)

// AddBuildingPreview godoc
// @Summary Заменить превью здания
// @Description Загружает новое изображение превью для здания
// @Tags buildings
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "ID здания (UUID)"
// @Param file formData file true "Изображение для превью"
// @Success 200 "OK"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings/{id}/image [post]
// @Security ApiKeyAuth
func AddBuildingPreview(
	log *slog.Logger,
	buildingsService *buildService.BuildingService,
) func(
	w http.ResponseWriter,
	r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("handling building preview upload request")

		// Проверяем ID здания
		id := r.PathValue("id")
		if err := uuid.Validate(id); err != nil {
			log.Error("invalid building ID", sl.Err(err))
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Invalid building ID",
			)
			return
		}
		// Проверяем права доступа

		// Парсим multipart форму
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
			log.Error("failed to parse multipart form", sl.Err(err))
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Failed to parse form",
			)
			return
		}

		files := r.MultipartForm.File["file"]
		if len(files) != 1 {
			log.Error("invalid number of files", "count", len(files))
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Exactly one file required",
			)
			return
		}

		file, err := files[0].Open()
		if err != nil {
			log.Error("failed to open uploaded file", sl.Err(err))
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Failed to process file",
			)
			return
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			log.Error("failed to read file contents", sl.Err(err))
			http_api.HandleError(
				w,
				http.StatusBadRequest,
				"Failed to read file",
			)
			return
		}

		log.Info(
			"uploading building preview",
			"buildingID",
			id,
			"fileSize",
			len(fileBytes),
		)

		if err := buildingsService.EditBuildingPreview(
			r.Context(),
			"",
			id,
			fileBytes,
		); err != nil {
			log.Error("failed to save building preview", sl.Err(err))
			http_api.HandleError(
				w,
				http.StatusInternalServerError,
				"Failed to save preview",
			)
			return
		}

		log.Info("building preview uploaded successfully", "buildingID", id)
		w.WriteHeader(http.StatusOK)
	}
}
