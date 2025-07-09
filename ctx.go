package nuke

import "net/http"

func CurrentUserID(r *http.Request) string {
	return r.Context().Value("user_id").(string)
}

func CurrentUserEmail(r *http.Request) string {
	return r.Context().Value("email").(string)
}

func CurrentUserUsername(r *http.Request) string {
	return r.Context().Value("username").(string)
}
