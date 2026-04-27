package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/notemg/notemg/internal/config"
	"github.com/notemg/notemg/internal/model"
	"github.com/notemg/notemg/internal/search"
	"github.com/notemg/notemg/internal/httputil"
	"github.com/notemg/notemg/internal/store"
)

type NoteHandler struct {
	noteSt   *store.NoteStore
	searcher *search.Searcher
	cfg      *config.Config
}

func NewNoteHandler(noteSt *store.NoteStore, searcher *search.Searcher, cfg *config.Config) *NoteHandler {
	return &NoteHandler{noteSt: noteSt, searcher: searcher, cfg: cfg}
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	params := model.NoteListParams{
		Page:     intQuery(q, "page", 1),
		PageSize: intQuery(q, "page_size", 50),
		SortBy:   q.Get("sort_by"),
		SortDesc: q.Get("sort_desc") == "true",
		Search:   q.Get("search"),
	}

	if fid := q.Get("folder_id"); fid != "" {
		params.FolderID = &fid
	}
	if tid := q.Get("tag_id"); tid != "" {
		params.TagID = &tid
	}
	if q.Get("is_deleted") == "true" {
		params.IsDeleted = true
	}

	result, err := h.noteSt.List(params)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	jsonResp(w, http.StatusOK, model.OK(result))
}

func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	note, err := h.noteSt.GetByID(id)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	if note == nil {
		jsonResp(w, http.StatusNotFound, model.Err(404, "Note not found"))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(note))
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input model.NoteCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	note, err := h.noteSt.Create(input)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	if h.searcher != nil {
		h.searcher.Index(note.ID, note.Title, note.Content)
	}

	jsonResp(w, http.StatusCreated, model.OK(note))
}

func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")

	var input model.NoteUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	note, err := h.noteSt.Update(id, input)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	if h.searcher != nil && note != nil {
		h.searcher.Index(note.ID, note.Title, note.Content)
	}

	jsonResp(w, http.StatusOK, model.OK(note))
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	if err := h.noteSt.Delete(id); err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(nil))
}

func (h *NoteHandler) Move(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	var req struct {
		FolderID *string `json:"folder_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.noteSt.Move(id, req.FolderID); err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(nil))
}

func (h *NoteHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	note, err := h.noteSt.Duplicate(id)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusCreated, model.OK(note))
}

func (h *NoteHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	if err := h.noteSt.Restore(id); err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(nil))
}

func (h *NoteHandler) PermanentDelete(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	if err := h.noteSt.PermanentDelete(id); err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	if h.searcher != nil {
		h.searcher.Delete(id)
	}
	jsonResp(w, http.StatusOK, model.OK(nil))
}

func intQuery(q url.Values, key string, defaultVal int) int {
	if v := q.Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultVal
}
