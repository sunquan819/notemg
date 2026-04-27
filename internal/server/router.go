package server

import (
	"net/http"
	"strings"

	"github.com/notemg/notemg/internal/httputil"
)

type Middleware func(http.Handler) http.Handler

type Route struct {
	method  string
	path    string
	handler http.Handler
}

type Router struct {
	routes      []Route
	middlewares []Middleware
	prefix      string
}

func newRouter() *Router {
	return &Router{}
}

func (r *Router) Group(prefix string) *Router {
	return &Router{
		routes:      r.routes,
		middlewares: r.middlewares,
		prefix:      r.prefix + prefix,
	}
}

func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.addRoute("GET", path, handler)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.addRoute("POST", path, handler)
}

func (r *Router) Put(path string, handler http.HandlerFunc) {
	r.addRoute("PUT", path, handler)
}

func (r *Router) Delete(path string, handler http.HandlerFunc) {
	r.addRoute("DELETE", path, handler)
}

func (r *Router) addRoute(method, path string, handler http.HandlerFunc) {
	fullPath := r.prefix + path
	h := http.Handler(handler)
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	r.routes = append(r.routes, Route{method: method, path: fullPath, handler: h})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	for _, route := range r.routes {
		if route.method != req.Method {
			continue
		}

		if matched, params := matchPath(route.path, path); matched {
			if len(params) > 0 {
				req = httputil.WithParams(req, params)
			}
			route.handler.ServeHTTP(w, req)
			return
		}
	}

	errJSON := `{"success":false,"error":{"code":404,"message":"Not found"}}`
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(errJSON))
}

func matchPath(pattern, path string) (bool, map[string]string) {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		if len(patternParts) == 0 || patternParts[len(patternParts)-1] != "*" {
			return false, nil
		}
		if len(pathParts) < len(patternParts)-1 {
			return false, nil
		}
	}

	params := make(map[string]string)
	for i := 0; i < len(patternParts); i++ {
		if i >= len(pathParts) {
			return false, nil
		}

		if patternParts[i] == "*" {
			params["*"] = strings.Join(pathParts[i:], "/")
			return true, params
		}

		if strings.HasPrefix(patternParts[i], "{") && strings.HasSuffix(patternParts[i], "}") {
			paramName := patternParts[i][1 : len(patternParts[i])-1]
			params[paramName] = pathParts[i]
			continue
		}

		if patternParts[i] != pathParts[i] {
			return false, nil
		}
	}

	return true, params
}