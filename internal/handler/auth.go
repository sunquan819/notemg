package handler

import (
	"encoding/json"
	"net/http"

	"github.com/notemg/notemg/internal/config"
	"github.com/notemg/notemg/internal/model"
	"github.com/notemg/notemg/internal/security"
)

type AuthHandler struct {
	auth *security.Auth
	cfg  *config.Config
}

func NewAuthHandler(auth *security.Auth, cfg *config.Config) *AuthHandler {
	return &AuthHandler{auth: auth, cfg: cfg}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ip := r.RemoteAddr
	if err := h.auth.CheckLoginAttempts(ip); err != nil {
		jsonResp(w, http.StatusTooManyRequests, model.Err(429, err.Error()))
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	if err := h.auth.VerifyPassword(h.cfg.Data.Dir, req.Password); err != nil {
		h.auth.RecordFailedAttempt(ip)
		jsonResp(w, http.StatusUnauthorized, model.Err(401, "Invalid password"))
		return
	}

	h.auth.ResetLoginAttempts(ip)
	access, refresh, err := h.auth.GenerateToken("user")
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Failed to generate token"))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    access,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.cfg.Auth.TokenExpire.Seconds()),
	})

	jsonResp(w, http.StatusOK, model.OK(map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	}))
}

func (h *AuthHandler) Init(w http.ResponseWriter, r *http.Request) {
	if h.auth.HasPassword(h.cfg.Data.Dir) {
		jsonResp(w, http.StatusConflict, model.Err(409, "Password already set"))
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	if len(req.Password) < 6 {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Password must be at least 6 characters"))
		return
	}

	if err := h.auth.SavePassword(h.cfg.Data.Dir, req.Password); err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Failed to save password"))
		return
	}

	access, refresh, _ := h.auth.GenerateToken("user")
	jsonResp(w, http.StatusOK, model.OK(map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	}))
}

func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, http.StatusOK, model.OK(map[string]bool{
		"initialized": h.auth.HasPassword(h.cfg.Data.Dir),
	}))
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	claims, err := h.auth.ValidateToken(req.RefreshToken, "refresh")
	if err != nil {
		jsonResp(w, http.StatusUnauthorized, model.Err(401, "Invalid refresh token"))
		return
	}

	access, refresh, _ := h.auth.GenerateToken(claims.Subject)
	jsonResp(w, http.StatusOK, model.OK(map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	}))
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Invalid request body"))
		return
	}

	if err := h.auth.VerifyPassword(h.cfg.Data.Dir, req.OldPassword); err != nil {
		jsonResp(w, http.StatusUnauthorized, model.Err(401, "Invalid current password"))
		return
	}

	if len(req.NewPassword) < 6 {
		jsonResp(w, http.StatusBadRequest, model.Err(400, "Password must be at least 6 characters"))
		return
	}

	if err := h.auth.SavePassword(h.cfg.Data.Dir, req.NewPassword); err != nil {
		jsonResp(w, http.StatusInternalServerError, model.Err(500, "Failed to save password"))
		return
	}

	jsonResp(w, http.StatusOK, model.OK(nil))
}

func jsonResp(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
