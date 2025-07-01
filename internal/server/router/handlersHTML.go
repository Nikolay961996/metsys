package router

import (
	"fmt"
	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"html/template"
	"net/http"
)

func getDashboardHandler(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := storage.GetAll()

		t, err := template.ParseFiles("./internal/server/router/metrics.html")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if err := t.Execute(w, metrics); err != nil {
			http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
