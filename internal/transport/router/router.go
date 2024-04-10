package router

import (
	"github.com/gorilla/mux"
)

type Router struct {
	router *mux.Router
}

type Handlify interface {
	Handle(*mux.Router)
}

func New(
	pathPrefix string,
) Router {

	r := mux.NewRouter().
		StrictSlash(false).
		PathPrefix(pathPrefix).
		Subrouter()

	return Router{
		router: r,
	}
}

func (r Router) Handle(
	routes map[string]Handlify,
) {

	for path, handler := range routes {
		hRouter := r.router.PathPrefix(path).Subrouter()
		handler.Handle(hRouter)
	}
}

func (r Router) Router() *mux.Router { return r.router }
