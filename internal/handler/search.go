package handler

import (
	"encoding/json"
	"net/http"

	"github.com/notemg/notemg/internal/model"
	"github.com/notemg/notemg/internal/search"
)

type SearchHandler struct {
	searcher *search.Searcher
}

func NewSearchHandler(searcher *search.Searcher) *SearchHandler {
	return &SearchHandler{searcher: searcher}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Search query is required"))
		return
	}

	if h.searcher == nil {
		jsonResp(w, http.StatusServiceUnavailable, model.Err(503, "Search not available"))
		return
	}

	results, err := h.searcher.Search(q)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	jsonResp(w, http.StatusOK, model.OK(results))
}

func (h *SearchHandler) Reindex(w http.ResponseWriter, r *http.Request) {
	if h.searcher == nil {
		jsonResp(w, http.StatusServiceUnavailable, model.Err(503, "Search not available"))
		return
	}

	var req struct {
		NoteIDs []string `json:"note_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	jsonResp(w, http.StatusOK, model.OK(map[string]int{"count": len(req.NoteIDs)}))
}
