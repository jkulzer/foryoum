package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jkulzer/foryoum/v2/controllers"
	"github.com/jkulzer/foryoum/v2/db"
	"github.com/jkulzer/foryoum/v2/routes"
)

func main() {
	port := 3000

	env := db.Init()
	controllers.ClearOutExpiredSessions(env)

	fmt.Println("Listening on :" + strconv.Itoa(port))
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	routes.Router(r, env)

	http.ListenAndServe(":"+strconv.Itoa(port), r)

	ticker := time.NewTicker(1 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				controllers.ClearOutExpiredSessions(env)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
