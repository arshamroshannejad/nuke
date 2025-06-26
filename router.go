package nuke

import (
	"net/http"
	"slices"
)

type Router struct {
	globalChain []func(handler http.Handler) http.Handler
	routeChain  []func(handler http.Handler) http.Handler
	isSubRouter bool
	mux         *http.ServeMux
}

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func (r *Router) Use(mw ...func(handler http.Handler) http.Handler) {
	if r.isSubRouter {
		r.routeChain = append(r.routeChain, mw...)
	} else {
		r.globalChain = append(r.globalChain, mw...)
	}
}

func (r *Router) Group(fn func(r *Router)) {
	subRouter := &Router{
		routeChain:  slices.Clone(r.routeChain),
		isSubRouter: true,
		mux:         r.mux,
	}
	fn(subRouter)
}

func (r *Router) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	r.Handle(pattern, handlerFunc)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	for _, mw := range slices.Backward(r.routeChain) {
		handler = mw(handler)
	}
	r.mux.Handle(pattern, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	var handler http.Handler = r.mux
	for _, mw := range slices.Backward(r.globalChain) {
		handler = mw(handler)
	}
	handler.ServeHTTP(w, rq)
}
