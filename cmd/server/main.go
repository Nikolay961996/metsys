package main

import (
	"github.com/Nikolay961996/metsys/internal/server/handlers"
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	metricsStorage := storage.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handlers.UpdateMetricHandler(metricsStorage))

	return http.ListenAndServe(":8080", mux)
}
