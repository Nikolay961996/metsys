package main

import (
	"github.com/Nikolay961996/metsys/internal/server/router"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":8080", router.MetricsRouter())
	if err != nil {
		panic(err)
	}
}
