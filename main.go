package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jkulzer/foryoum/v2/db"
	"github.com/jkulzer/foryoum/v2/routes"
)

func main() {
	port := 3000

	env := db.Init()

	fmt.Println("Listening on :" + strconv.Itoa(port))
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	routes.Router(r, env)

	http.ListenAndServe(":"+strconv.Itoa(port), r)
}
