package api

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Middleware func(http.Handler) http.Handler

type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func (r *Router) Use(mw ...Middleware) {
	r.middlewares = append(r.middlewares, mw...)
}

func (r *Router) HandleFunc(pattern string, fn http.HandlerFunc) {
	var h http.Handler = fn
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}

	r.mux.Handle(pattern, h)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func InitRoutes(db *pgxpool.Pool) http.Handler {
	r := NewRouter()

	//TODO: add middleware layyers

	//TODO: other api handlers endpoints

	return r
}
