package httputil

import (
	"context"
	"net/http"
)

type ctxKey string

const paramsKey ctxKey = "params"

func PathParam(r *http.Request, key string) string {
	if params, ok := r.Context().Value(paramsKey).(map[string]string); ok {
		return params[key]
	}
	return ""
}

func WithParams(r *http.Request, params map[string]string) *http.Request {
	ctx := context.WithValue(r.Context(), paramsKey, params)
	return r.WithContext(ctx)
}
