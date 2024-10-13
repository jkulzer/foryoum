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

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

func Router(r chi.Router, env *db.Env, customContent string) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(views.Main(customContent)).ServeHTTP(w, r)
	})

	r.Route("/posts", func(r chi.Router) {
		r.Get("/{index}",
			func(w http.ResponseWriter, r *http.Request) {
				index, err := strconv.ParseUint(chi.URLParam(r, "index"), 10, 0)
				if err != nil {
					templ.Handler(views.GenericError("Invalid Post range", customContent)).ServeHTTP(w, r)
				}
				isLoggedIn, _ := controllers.GetLoginFromSession(env, r)

				posts, lastPage := controllers.GetPostList(env, uint(index))
				templ.Handler(views.PostView(posts, index, lastPage, isLoggedIn, customContent)).ServeHTTP(w, r)
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
				templ.Handler(views.SearchPage(customContent)).ServeHTTP(w, r)
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
		r.Post("/preview",
			func(w http.ResponseWriter, r *http.Request) {
				response, err := helpers.ReadHttpResponse(r.Body)
				if err != nil {
					fmt.Println("Failed to read HTTP response")
				}

				data, err := url.ParseQuery(response)
				if err != nil {
					fmt.Println("Failed to parse query")
				}
				input := data["body"][0]
				unsafe := blackfriday.MarkdownCommon([]byte(input))
				html := string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
				fmt.Fprint(w, "<div id=\"post-preview\">"+html+"</div>")
			},
		)
		r.Get("/{postId}",
			func(w http.ResponseWriter, r *http.Request) {
				// 10 is base 10 and 0 indicates parsing into system-size int
				postId, err := strconv.ParseUint(chi.URLParam(r, "postId"), 10, 0)
				if err != nil {
					templ.Handler(views.GenericError("Invalid Post ID", customContent)).ServeHTTP(w, r)
				}
				templ.Handler(views.RedirectTo("post/"+fmt.Sprint(postId)+"/0")).ServeHTTP(w, r)
			},
		)
		r.Get("/{postId}/{commentIndex}",
			func(w http.ResponseWriter, r *http.Request) {
				// 10 is base 10 and 0 indicates parsing into system-size int
				postId, err := strconv.ParseUint(chi.URLParam(r, "postId"), 10, 0)
				if err != nil {
					templ.Handler(views.GenericError("Invalid Post ID", customContent)).ServeHTTP(w, r)
				}

				var post models.RootPost
				result := env.DB.First(&post, postId)
				if result.Error != nil {
					templ.Handler(views.GenericError("Failed to load posts", customContent)).ServeHTTP(w, r)
				} else {

					comments := controllers.GetCommentList(env, uint(postId))
					isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
					templ.Handler(views.Post(post, comments, customContent, isLoggedIn)).ServeHTTP(w, r)
				}
			},
		)
	})
	r.Route("/comment", func(r chi.Router) {
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

					postId, err := strconv.ParseUint(data["rootPostID"][0], 10, 0)
					if err != nil {
						templ.Handler(views.GenericError("Invalid Post ID", customContent)).ServeHTTP(w, r)
					}

					fmt.Println("Creating a comment with root post id \"" + fmt.Sprint(postId) + "\" and body \"" + data["text"][0])
					env.DB.Create(&models.Comment{
						RootPostID:   uint(postId),
						Body:         data["text"][0],
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
	})
	r.Route("/register", func(r chi.Router) {
		r.Get("/",
			func(w http.ResponseWriter, r *http.Request) {
				templ.Handler(views.Register(customContent)).ServeHTTP(w, r)
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
				templ.Handler(views.Login(redirect, customContent)).ServeHTTP(w, r)
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
				templ.Handler(views.Logout(isLoggedIn, customContent)).ServeHTTP(w, r)
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

				templ.Handler(views.SessionList(isLoggedIn, sessionList, customContent)).ServeHTTP(w, r)
			},
		)
		r.Delete("/{session}",
			func(w http.ResponseWriter, r *http.Request) {
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
