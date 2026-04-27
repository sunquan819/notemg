package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/notemg/notemg/internal/config"
	"github.com/notemg/notemg/internal/model"
	"github.com/notemg/notemg/internal/httputil"
)

type ImageHandler struct {
	cfg *config.Config
}

func NewImageHandler(cfg *config.Config) *ImageHandler {
	return &ImageHandler{cfg: cfg}
}

func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "File too large (max 50MB)"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "No file provided"))
		return
	}
	defer file.Close()

	id := uuid.New().String()
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".bin"
	}

	storedName := id + ext
	storedPath := filepath.Join(h.cfg.Data.AttachmentsDir, storedName)

	dst, err := os.Create(h.cfg.DataDir(storedPath))
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Failed to save file"))
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Failed to save file"))
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	att := model.Attachment{
		ID:        id,
		Filename:  header.Filename,
		FilePath:  storedPath,
		FileSize:  written,
		MimeType:  mimeType,
	}

	jsonResp(w, http.StatusCreated, model.OK(att))
}

func (h *ImageHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")

	attachmentsDir := h.cfg.AttachmentsPath()
	entries, err := os.ReadDir(attachmentsDir)
	if err != nil {
		jsonResp(w, http.StatusNotFound, model.Err(404, "Attachment not found"))
		return
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), id+".") {
			filePath := filepath.Join(attachmentsDir, entry.Name())
			ext := filepath.Ext(entry.Name())
			contentType := mimeTypeFromExt(ext)
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			http.ServeFile(w, r, filePath)
			return
		}
	}

	jsonResp(w, http.StatusNotFound, model.Err(404, "Attachment not found"))
}

func (h *ImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")

	attachmentsDir := h.cfg.AttachmentsPath()
	entries, _ := os.ReadDir(attachmentsDir)

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), id+".") {
			os.Remove(filepath.Join(attachmentsDir, entry.Name()))
			jsonResp(w, http.StatusOK, model.OK(nil))
			return
		}
	}

	jsonResp(w, http.StatusNotFound, model.Err(404, "Attachment not found"))
}

func mimeTypeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}
