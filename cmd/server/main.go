package main

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/router"
)

func main() {
	h := handler.New()
	r := router.New(h)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
