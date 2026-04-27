package handler

import (
	"encoding/json"
	"net/http"

	"github.com/notemg/notemg/internal/model"
	"github.com/notemg/notemg/internal/httputil"
	"github.com/notemg/notemg/internal/store"
)

type FolderHandler struct {
	folderSt *store.FolderStore
}

func NewFolderHandler(folderSt *store.FolderStore) *FolderHandler {
	return &FolderHandler{folderSt: folderSt}
}

func (h *FolderHandler) Tree(w http.ResponseWriter, r *http.Request) {
	folders, err := h.folderSt.Tree()
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(folders))
}

func (h *FolderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input model.FolderCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	if input.Name == "" {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Folder name is required"))
		return
	}

	folder, err := h.folderSt.Create(input)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	jsonResp(w, http.StatusCreated, model.OK(folder))
}

func (h *FolderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")

	var input model.FolderUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	folder, err := h.folderSt.Update(id, input)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}

	jsonResp(w, http.StatusOK, model.OK(folder))
}

func (h *FolderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := httputil.PathParam(r, "id")
	if err := h.folderSt.Delete(id); err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, err.Error()))
		return
	}
	jsonResp(w, http.StatusOK, model.OK(nil))
}
