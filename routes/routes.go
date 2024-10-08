package routes

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
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
		templ.Handler(views.Main()).ServeHTTP(w, r)
	})

	r.Route("/posts", func(r chi.Router) {
		r.Get("/{index}",
			func(w http.ResponseWriter, r *http.Request) {
				index, err := strconv.ParseUint(chi.URLParam(r, "index"), 10, 0)
				if err != nil {
					templ.Handler(views.GenericError("Invalid Post range")).ServeHTTP(w, r)
				}
				isLoggedIn, _ := controllers.GetLoginFromSession(env, r)

				posts, lastPage := controllers.GetPostList(env, uint(index))
				templ.Handler(views.PostView(posts, index, lastPage, isLoggedIn)).ServeHTTP(w, r)
			},
		)
		r.Get("/",
			func(w http.ResponseWriter, r *http.Request) {
				templ.Handler(views.PostRedirect()).ServeHTTP(w, r)
			},
		)
	})
	r.Route("/search", func(r chi.Router) {
		r.Get("/",
			func(w http.ResponseWriter, r *http.Request) {
				templ.Handler(views.SearchPage()).ServeHTTP(w, r)
			},
		)
		r.Post("/",
			func(w http.ResponseWriter, r *http.Request) {
				response, err := helpers.ReadHttpResponse(r.Body)
				if err != nil {
					fmt.Println("Failed to read HTTP response")
				}
				index := uint64(0)
				data, err := url.ParseQuery(response)
				if err != nil {
					fmt.Println("Failed to parse query")
				}

				searchTerm := data["searchTerm"][0]
				fmt.Println("searching for " + searchTerm)
				posts, lastPage := controllers.SearchPostList(env, searchTerm)
				templ.Handler(views.SearchResults(posts, index, lastPage)).ServeHTTP(w, r)
			},
		)
	})
	r.Route("/post", func(r chi.Router) {
		r.Post("/",
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
				if isLoggedIn {
					currentTime := time.Now()

					env.DB.Create(&models.RootPost{
						Title:        data["title"][0],
						Body:         data["body"][0],
						CreationDate: currentTime,
						UpdateDate:   currentTime,
						Op:           session.UserAccount.Name,
						Version:      1,
					})
					// gets a new session token
					controllers.RefreshSession(env, w, r)
				}
			},
		)
		r.Get("/{postId}",
			func(w http.ResponseWriter, r *http.Request) {
				// 10 is base 10 and 0 indicates parsing into system-size int
				postId, err := strconv.ParseUint(chi.URLParam(r, "postId"), 10, 0)
				if err != nil {
					templ.Handler(views.GenericError("Invalid Post ID")).ServeHTTP(w, r)
				}
				var post models.RootPost
				result := env.DB.First(&post, postId)
				if result.Error != nil {
					templ.Handler(views.GenericError("Failed to load posts")).ServeHTTP(w, r)
				} else {
					templ.Handler(views.Post(post)).ServeHTTP(w, r)
				}
			},
		)
	})
	r.Route("/register", func(r chi.Router) {
		r.Get("/",
			func(w http.ResponseWriter, r *http.Request) {
				templ.Handler(views.Register()).ServeHTTP(w, r)
			},
		)

		r.Post("/",
			func(w http.ResponseWriter, r *http.Request) {
				response, err := helpers.ReadHttpResponse(r.Body)
				data, err := url.ParseQuery(response)
				if err != nil {
					fmt.Println("Failed to parse query")
				}

				password := data["password"][0]

				hashedPassword, err := controllers.HashPassword(password)
				if err != nil {
					fmt.Println("Failed to hash password")
				}

				userName := models.UserAccount{
					Name:     data["username"][0],
					Password: hashedPassword,
				}
				// tries to create the user in the db
				result := env.DB.Create(&userName)

				// if the user creation fails,
				if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
					fmt.Println("Duplicate Username")
					templ.Handler(views.RegistrationFailed()).ServeHTTP(w, r)
				} else {
					controllers.CreateSession(env, userName, w)
				}
			},
		)
	})
	r.Route("/login", func(r chi.Router) {
		r.Get("/*",
			func(w http.ResponseWriter, r *http.Request) {
				redirect := chi.URLParam(r, "*")
				templ.Handler(views.Login(redirect)).ServeHTTP(w, r)
			},
		)
		r.Post("/",
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
				redirect := data["redirect"][0]

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
						controllers.CreateSession(env, userAccount, w)
						fmt.Println("Redirecting to " + redirect)
						templ.Handler(views.RedirectTo(redirect)).ServeHTTP(w, r)
					} else {
						templ.Handler(views.WrongPassword()).ServeHTTP(w, r)
					}
				}

			},
		)
	})
	r.Route("/logout", func(r chi.Router) {
		r.Get("/",
			func(w http.ResponseWriter, r *http.Request) {
				isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
				templ.Handler(views.Logout(isLoggedIn)).ServeHTTP(w, r)
			},
		)
		r.Post("/",
			func(w http.ResponseWriter, r *http.Request) {
				controllers.RemoveSession(env, w, r)
			},
		)
	})
	r.Route("/sessions", func(r chi.Router) {
		r.Get("/",
			func(w http.ResponseWriter, r *http.Request) {
				isLoggedIn, session := controllers.GetLoginFromSession(env, r)

				sessionList := controllers.GetSessionsForUser(env, r, session)

				templ.Handler(views.SessionList(isLoggedIn, sessionList)).ServeHTTP(w, r)
			},
		)
		r.Delete("/{session}",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("Trying to delete")
				sessionTokenString := chi.URLParam(r, "session")
				sessionToken, err := uuid.Parse(sessionTokenString)
				if err != nil {
					fmt.Println("Failed to parse UUID")
				}
				controllers.DeleteSessionByUuid(sessionToken, env, r)
			},
		)
	})
}
