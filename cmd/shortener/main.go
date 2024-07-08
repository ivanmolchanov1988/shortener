package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	//"github.com/go-chi/chi/v5/middleware"
	"github.com/ivanmolchanov1988/shortener/pkg/handlers"
	"github.com/ivanmolchanov1988/shortener/pkg/memory"
	//"github.com/ivanmolchanov1988/shortener/pkg/server"
)

func main() {

	memStore := memory.NewMemoryStorage()
	handler := handlers.NewHandler(memStore)
	//srv := server.NewServer(handler)

	// err := srv.ListenAndServe()
	// if err != nil {
	// 	panic(err)
	// }

	r := chi.NewRouter()
	r.Post("/", handler.PostUrl)
	r.Get("/{id}", handler.GetUrl)

	http.ListenAndServe(":8080", r)
}
