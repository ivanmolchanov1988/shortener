package main

import (
	"github.com/ivanmolchanov1988/shortener/pkg/handlers"
	"github.com/ivanmolchanov1988/shortener/pkg/memory"
	"github.com/ivanmolchanov1988/shortener/pkg/server"
)

func main() {

	memStore := memory.NewMemoryStorage()
	handler := handlers.NewHandler(memStore)
	srv := server.NewServer(handler)

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
