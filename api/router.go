package api

import "net/http"

type Router struct {
	*http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

func NewRouter() *Router {
	return &Router{ServeMux: http.NewServeMux()}
}

func (r *Router) Use(m func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, m)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}
	r.ServeMux.Handle(pattern, handler)
}

func (r *Router) HandleFunc(pattern string, fn func(w http.ResponseWriter, r *http.Request)) {
	r.Handle(pattern, http.HandlerFunc(fn))
}
