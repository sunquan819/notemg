package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/notemg/notemg/internal/config"
	"github.com/notemg/notemg/internal/httputil"
	"github.com/notemg/notemg/internal/model"
	"github.com/notemg/notemg/internal/store"
)

type DriveHandler struct {
	driveSt *store.DriveStore
	cfg     *config.Config
}

func NewDriveHandler(driveSt *store.DriveStore, cfg *config.Config) *DriveHandler {
	return &DriveHandler{driveSt: driveSt, cfg: cfg}
}

func (h *DriveHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	params := model.DriveListParams{
		Type:     q.Get("type"),
		Search:   q.Get("search"),
		SortBy:   q.Get("sort_by"),
		SortDesc: q.Get("sort_desc") == "true",
	}
	if pid := q.Get("parent_id"); pid != "" {
		params.ParentID = &pid
	}

	result, err := h.driveSt.List(params)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	jsonResp(w, http.StatusOK, model.OK(result))
}

func (h *DriveHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	file, err := h.driveSt.GetByID(id)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	if file == nil {
		jsonResp(w, http.StatusNotFound, model.Err(404, "File not found"))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(file))
}

func (h *DriveHandler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request"))
		return
	}

	if req.Name == "" {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Name required"))
		return
	}

	folder, err := h.driveSt.CreateFolder(req.Name, req.ParentID)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	jsonResp(w, http.StatusCreated, model.OK(folder))
}

func (h *DriveHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 500<<20)

	if err := r.ParseMultipartForm(500 << 20); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "File too large (max 500MB)"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "No file provided"))
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Read file failed"))
		return
	}

	parentID := r.FormValue("parent_id")
	var pid *string
	if parentID != "" {
		pid = &parentID
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = guessMimeType(header.Filename)
	}

	driveFile, err := h.driveSt.SaveFile(header.Filename, mimeType, int64(len(content)), pid, content)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	jsonResp(w, http.StatusCreated, model.OK(driveFile))
}

func (h *DriveHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	file, err := h.driveSt.GetByID(id)
	if err != nil || file == nil {
		jsonResp(w, http.StatusNotFound, model.Err(404, "File not found"))
		return
	}

	if file.Type != "file" {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Not a file"))
		return
	}

	data, err := os.ReadFile(file.Path)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Read file failed"))
		return
	}

	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+sanitizeFilename(file.Name)+"\"")
	w.Header().Set("Content-Length", string(len(data)))
	w.Write(data)
}

func (h *DriveHandler) Preview(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	file, err := h.driveSt.GetByID(id)
	if err != nil || file == nil {
		jsonResp(w, http.StatusNotFound, model.Err(404, "File not found"))
		return
	}

	if file.Type != "file" {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Not a file"))
		return
	}

	data, err := os.ReadFile(file.Path)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Read file failed"))
		return
	}

	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Length", string(len(data)))
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(data)
}

func (h *DriveHandler) Thumbnail(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	file, err := h.driveSt.GetByID(id)
	if err != nil || file == nil {
		jsonResp(w, http.StatusNotFound, model.Err(404, "File not found"))
		return
	}

	if file.ThumbPath == "" {
		h.Preview(w, r)
		return
	}

	data, err := os.ReadFile(file.ThumbPath)
	if err != nil {
		h.Preview(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(data)
}

func (h *DriveHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	if err := h.driveSt.Delete(id); err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(nil))
}

func (h *DriveHandler) Rename(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request"))
		return
	}

	file, err := h.driveSt.Rename(id, req.Name)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(file))
}

func (h *DriveHandler) Move(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	var req struct {
		ParentID *string `json:"parent_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	file, err := h.driveSt.Move(id, req.ParentID)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(file))
}

func (h *DriveHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Query required"))
		return
	}

	files, err := h.driveSt.Search(q)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	jsonResp(w, http.StatusOK, model.OK(files))
}

func guessMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".svg":
		return "image/svg+xml"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".zip":
		return "application/zip"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}