package routes

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jkulzer/foryoum/v2/controllers"
	"github.com/jkulzer/foryoum/v2/db"
	"github.com/jkulzer/foryoum/v2/helpers"
	"github.com/jkulzer/foryoum/v2/models"
	"github.com/jkulzer/foryoum/v2/views"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

func Router(r chi.Router, env *db.Env) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		posts := controllers.GetPostList(25, env, 0)
		templ.Handler(views.Main(posts)).ServeHTTP(w, r)
	})

	r.Post("/post",
		func(w http.ResponseWriter, r *http.Request) {
			response, err := helpers.ReadHttpResponse(r.Body)
			if err != nil {
				fmt.Println("Failed to read HTTP response")
			}

			data, err := url.ParseQuery(response)
			if err != nil {
				fmt.Println("Failed to parse query")
			}

			currentTime := time.Now()

			env.DB.Create(&models.RootPost{
				Title:        data["title"][0],
				Body:         data["body"][0],
				CreationDate: currentTime,
				UpdateDate:   currentTime,
				Op:           "meee :3",
				Version:      1,
			})
			fmt.Println(response)
		},
	)

	r.Get("/posts",
		func(w http.ResponseWriter, r *http.Request) {

			posts := controllers.GetPostList(25, env, 0)

			templ.Handler(views.PostList(posts)).ServeHTTP(w, r)
		},
	)

}
