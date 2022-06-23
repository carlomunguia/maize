package main

import "net/http"

// SessionLoad is a middleware that loads the session from the request.
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}
