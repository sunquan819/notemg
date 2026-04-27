package handler

import (
	"encoding/json"
	"net/http"

	"github.com/notemg/notemg/internal/markdown"
	"github.com/notemg/notemg/internal/model"
)

type MarkdownHandler struct {
	md *markdown.Engine
}

func NewMarkdownHandler(md *markdown.Engine) *MarkdownHandler {
	return &MarkdownHandler{md: md}
}

func (h *MarkdownHandler) Render(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	html, err := h.md.Render(req.Content)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Failed to render markdown"))
		return
	}

	jsonResp(w, http.StatusOK, model.OK(map[string]string{"html": html}))
}
