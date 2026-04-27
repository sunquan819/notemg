package handler

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/notemg/notemg/internal/config"
	"github.com/notemg/notemg/internal/markdown"
	"github.com/notemg/notemg/internal/model"
	"github.com/notemg/notemg/internal/httputil"
	"github.com/notemg/notemg/internal/store"
)

type ImportExportHandler struct {
	noteSt *store.NoteStore
	md     *markdown.Engine
	cfg    *config.Config
}

func NewImportExportHandler(noteSt *store.NoteStore, md *markdown.Engine, cfg *config.Config) *ImportExportHandler {
	return &ImportExportHandler{noteSt: noteSt, md: md, cfg: cfg}
}

func (h *ImportExportHandler) ImportMarkdown(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "File too large"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "No file provided"))
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".md") {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Only .md files are supported"))
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Failed to read file"))
		return
	}

	title := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	folderID := r.FormValue("folder_id")
	var fid *string
	if folderID != "" {
		fid = &folderID
	}

	note, err := h.noteSt.Create(model.NoteCreate{
		Title:    title,
		FolderID: fid,
		Content:  string(content),
	})
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	jsonResp(w, http.StatusCreated, model.OK(note))
}

func (h *ImportExportHandler) ImportZip(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "File too large"))
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "No file provided"))
		return
	}
	defer file.Close()

	tmpFile, err := os.CreateTemp("", "notemg-import-*.zip")
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Failed to create temp file"))
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, file); err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Failed to save temp file"))
		return
	}

	zipReader, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid ZIP file"))
		return
	}
	defer zipReader.Close()

	imported := make([]string, 0)
	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() || !strings.HasSuffix(strings.ToLower(f.Name), ".md") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}

		title := strings.TrimSuffix(filepath.Base(f.Name), ".md")
		note, err := h.noteSt.Create(model.NoteCreate{
			Title:   title,
			Content: string(content),
		})
		if err == nil && note != nil {
			imported = append(imported, note.ID)
		}
	}

	jsonResp(w, http.StatusOK, model.OK(map[string]interface{}{
		"imported": imported,
		"count":    len(imported),
	}))
}

func (h *ImportExportHandler) ExportNote(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "markdown"
	}

	note, err := h.noteSt.GetByID(id)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	if note == nil {
		jsonResp(w, http.StatusNotFound, model.Err(404, "Note not found"))
		return
	}

	switch format {
	case "markdown", "md":
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.md"`, sanitizeFilename(note.Title)))
		w.Write([]byte(note.Content))

	case "html":
		htmlContent, err := h.md.Render(note.Content)
		if err != nil {
			jsonResp(w, http.StatusInternalServerError, model.Err(500, "Failed to render"))
			return
		}
		fullHTML := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>%s</title><style>body{max-width:800px;margin:0 auto;padding:20px;font-family:system-ui,-apple-system,sans-serif;line-height:1.6}code{background:#f4f4f4;padding:2px 6px;border-radius:3px}pre{background:#f4f4f4;padding:16px;border-radius:6px;overflow-x:auto}img{max-width:100%%}table{border-collapse:collapse;width:100%%}th,td{border:1px solid #ddd;padding:8px 12px;text-align:left}</style></head><body>%s</body></html>`, note.Title, htmlContent)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.html"`, sanitizeFilename(note.Title)))
		w.Write([]byte(fullHTML))

	default:
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Unsupported format: "+format))
	}
}

func (h *ImportExportHandler) ExportBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NoteIDs []string `json:"note_ids"`
		Format  string   `json:"format"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	if len(req.NoteIDs) == 0 {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "No note IDs provided"))
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=notemg-export.zip")
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, id := range req.NoteIDs {
		note, err := h.noteSt.GetByID(id)
		if err != nil || note == nil {
			continue
		}

		filename := sanitizeFilename(note.Title) + ".md"
		fw, err := zipWriter.Create(filename)
		if err != nil {
			continue
		}
		fw.Write([]byte(note.Content))
	}
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, "\"", "-")
	if name == "" {
		name = "untitled"
	}
	return name
}
