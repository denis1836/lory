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
		mux:         http.NewServeMux(),
		middlewares: []Middleware{},
	}
}

func (r *Router) Use(mw ...Middleware) {
	r.middlewares = append(r.middlewares, mw...)
}

func (r *Router) Get(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	r.handle("GET", path, handler, middlewares...)
}

func (r *Router) Post(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	r.handle("POST", path, handler, middlewares...)
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

func (r *Router) handle(method, path string, handler http.Handler, localMiddlewares ...Middleware) {
	finalHandler := chain(handler, localMiddlewares...)

	finalHandler = chain(finalHandler, r.middlewares...)

	pattern := method + " " + path
	r.mux.Handle(pattern, finalHandler)
}

func chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
