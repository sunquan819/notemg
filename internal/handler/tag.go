package handler

import (
	"encoding/json"
	"net/http"

	"github.com/notemg/notemg/internal/model"
	"github.com/notemg/notemg/internal/httputil"
	"github.com/notemg/notemg/internal/store"
)

type TagHandler struct {
	tagSt *store.TagStore
}

func NewTagHandler(tagSt *store.TagStore) *TagHandler {
	return &TagHandler{tagSt: tagSt}
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	tags, err := h.tagSt.List()
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(tags))
}

func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input model.TagCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	if input.Name == "" {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Tag name is required"))
		return
	}

	tag, err := h.tagSt.Create(input)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	jsonResp(w, http.StatusCreated, model.OK(tag))
}

func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	if err := h.tagSt.Delete(id); err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(nil))
}

func (h *TagHandler) NotesByTag(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	ids, err := h.tagSt.NotesByTag(id)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(ids))
}
