package gowebi

import (
	"net/http"
	"path/filepath"
)

func ServeBundle(dir string) http.Handler {
	dir = filepath.Join(dir, "client")
	return http.StripPrefix("/client/", http.FileServer(http.Dir(dir)))
}
