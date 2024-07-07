package server

import (
	"net/http"
	"strings"

	"github.com/ivanmolchanov1988/shortener/pkg/handlers"
)

func forGet(path string) bool {
	return strings.HasPrefix(path, "/") && len(path) > 1
}

func NewServer(h *handlers.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/" {
			h.PostUrl(w, r)
		} else if r.Method == http.MethodGet && forGet(r.URL.Path) {
			h.GetUrl(w, r)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})

	return &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
}
