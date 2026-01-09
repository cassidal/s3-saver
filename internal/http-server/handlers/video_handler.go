package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"s3-saver/internal/config"
	logUtil "s3-saver/internal/lib/logger/slog"
	"s3-saver/internal/service"
)

type VideoHandler struct {
	storage service.StorageService
	log     slog.Logger
	config  config.AppConfig
	history service.HistoryService
}

func NewVideoHandler(storage service.StorageService, log *slog.Logger, appConfig *config.AppConfig, history service.HistoryService) *VideoHandler {
	return &VideoHandler{
		storage: storage,
		log:     *log,
		config:  *appConfig,
		history: history,
	}
}

// Upload загружает видео файл в S3 хранилище
// @Summary      Загрузить видео
// @Description  Загружает видео файл в S3 хранилище. Максимальный размер файла: 500 MB
// @Tags         video
// @Accept       multipart/form-data
// @Produce      json
// @Param        video  formData  file  true  "Видео файл для загрузки"
// @Success      200    {object}  map[string]string  "Успешная загрузка"
// @Failure      400    {string}  string             "Ошибка валидации (файл слишком большой или невалидный)"
// @Failure      500    {string}  string             "Внутренняя ошибка сервера"
// @Router       /upload/video [post]
func (h *VideoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 500 << 20 // 500 MB

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.log.Warn("failed to parse multipart form", "error", logUtil.Err(err))
		http.Error(w, "File too big", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		h.log.Warn("failed to get video file from form", "error", logUtil.Err(err))
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadedFileName, err := h.storage.UploadFile(r.Context(), file, header)
	if err != nil {
		fmt.Printf("Upload error: %v\n", err)
		http.Error(w, "Failed to upload video", http.StatusInternalServerError)
		return
	}

	fullUrl := fmt.Sprintf("%s/%s/%s", h.config.S3Config.Endpoint, h.config.S3Config.BucketName, uploadedFileName)
	h.history.Add(fullUrl)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":   "success",
		"filename": fullUrl,
	}); err != nil {
		h.log.Warn("JSON encoding error", "error", logUtil.Err(err))
		return
	}
	// Принудительно отправляем ответ, если поддерживается
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// GetRecentList возвращает список недавно загруженных видео файлов
// @Summary      Получить список недавно загруженных файлов
// @Description  Возвращает список URL последних загруженных видео файлов (до 5 файлов)
// @Tags         video
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Список недавно загруженных файлов"
// @Failure      500  {string}  string                  "Внутренняя ошибка сервера"
// @Router       /recent [get]
func (h *VideoHandler) GetRecentList(w http.ResponseWriter, r *http.Request) {
	urls := h.history.GetRecent()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recent_files": urls,
	})
}

// LatestVideoHandler возвращает ссылку на последнее загруженное видео через редирект
// @Summary      Получить последнее видео (редирект)
// @Description  Возвращает редирект (302) на URL самого свежего видео из S3 хранилища.
//
//	Идеально подходит для автоматизации скачивания без парсинга JSON.
//
// @Tags         video
// @Produce      text/plain
// @Success      302  {string}  string  "Редирект на S3 URL последнего видео"
// @Failure      404  {string}  string  "Нет доступных видео"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /video/latest [get]
func (h *VideoHandler) LatestVideoHandler(w http.ResponseWriter, r *http.Request) {
	urls := h.history.GetRecent()

	if len(urls) == 0 {
		http.Error(w, "No videos available", http.StatusNotFound)
		return
	}

	latestURL := urls[0]

	http.Redirect(w, r, latestURL, http.StatusFound)
}
