package routes

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"strconv"
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
		isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
		templ.Handler(views.Main(posts, isLoggedIn)).ServeHTTP(w, r)
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

			isLoggedIn, session := controllers.GetLoginFromSession(env, r)
			fmt.Println("Are they logged in? " + strconv.FormatBool(isLoggedIn))
			fmt.Println(session)
			if isLoggedIn {

				currentTime := time.Now()

				env.DB.Create(&models.RootPost{
					Title:        data["title"][0],
					Body:         data["body"][0],
					CreationDate: currentTime,
					UpdateDate:   currentTime,
					Op:           "test",
					Version:      1,
					// Op:           userAccount.Name,
				})
				fmt.Println("Username: " + session.UserAccount.Name)
			}
		},
	)

	r.Get("/posts",
		func(w http.ResponseWriter, r *http.Request) {

			posts := controllers.GetPostList(25, env, 0)

			templ.Handler(views.PostList(posts)).ServeHTTP(w, r)
		},
	)
	r.Get("/register",
		func(w http.ResponseWriter, r *http.Request) {
			templ.Handler(views.Register()).ServeHTTP(w, r)
		},
	)

	r.Post("/register",
		func(w http.ResponseWriter, r *http.Request) {
			response, err := helpers.ReadHttpResponse(r.Body)
			if err != nil {
				fmt.Println("Failed to read HTTP response")
			}

			data, err := url.ParseQuery(response)
			if err != nil {
				fmt.Println("Failed to parse query")
			}

			password := data["password"][0]

			hashedPassword, err := controllers.HashPassword(password)
			if err != nil {
				fmt.Println("Failed to hash password")
			}

			fmt.Println(data["username"][0])

			result := env.DB.Create(&models.UserAccount{
				Name:     data["username"][0],
				Password: hashedPassword,
			})

			if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
				fmt.Println("Duplicate Username")
				templ.Handler(views.RegistrationFailed()).ServeHTTP(w, r)
			}
		},
	)
	r.Get("/login",
		func(w http.ResponseWriter, r *http.Request) {
			templ.Handler(views.Login()).ServeHTTP(w, r)
		},
	)
	r.Post("/login",
		func(w http.ResponseWriter, r *http.Request) {
			response, err := helpers.ReadHttpResponse(r.Body)
			if err != nil {
				fmt.Println("Failed to read HTTP response")
			}

			data, err := url.ParseQuery(response)
			if err != nil {
				fmt.Println("Failed to parse query")
			}

			username := data["username"][0]
			password := data["password"][0]

			var userAccount models.UserAccount
			result := env.DB.Where(&models.UserAccount{Name: username}).First(&userAccount)

			if result.Error != nil {
				fmt.Println("Username not found")
				templ.Handler(views.UserNameNotFound()).ServeHTTP(w, r)
			} else {

				// checks if password is correct
				if controllers.CheckPasswordHash(
					password, userAccount.Password,
				) {
					fmt.Println(userAccount)
					sessionToken, expiryDuration := controllers.NewSession(env, userAccount)
					// creates a session cookie
					cookie := http.Cookie{
						Name:  "Session",
						Value: sessionToken,
						Path:  "/",
						// sets the expiry time also used in the session
						MaxAge:   int(expiryDuration.Seconds()),
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteLaxMode,
					}

					http.SetCookie(w, &cookie)
				} else {
					templ.Handler(views.WrongPassword()).ServeHTTP(w, r)
				}
			}

		},
	)
}
