package nuke

import "net/http"

func URLParam(r *http.Request, name string) string {
	return r.PathValue(name)
}

func QueryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}
