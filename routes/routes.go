package routes

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/jkulzer/foryoum/v2/controllers"
	"github.com/jkulzer/foryoum/v2/db"
	"github.com/jkulzer/foryoum/v2/helpers"
	"github.com/jkulzer/foryoum/v2/models"
	"github.com/jkulzer/foryoum/v2/translations"
	"github.com/jkulzer/foryoum/v2/views"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

func Router(r chi.Router, env *db.Env, customContent string, mainPage string) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		lang := translations.GetLanguageFromCookie(r)
		t := translations.GetTranslations(lang)
		isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
		templ.Handler(views.Main(mainPage, customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
	})
	r.Route("/posts", func(r chi.Router) {
		r.Get("/{index}",
			func(w http.ResponseWriter, r *http.Request) {
				index, err := strconv.ParseUint(chi.URLParam(r, "index"), 10, 0)
				isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
				if err != nil {
					lang := translations.GetLanguageFromCookie(r)
					t := translations.GetTranslations(lang)
					templ.Handler(views.GenericError(t.InvalidPostRange, customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
				}

				posts, lastPage := controllers.GetPostList(env, uint(index))
				lang := translations.GetLanguageFromCookie(r)
				t := translations.GetTranslations(lang)
				templ.Handler(views.PostView(posts, index, lastPage, isLoggedIn, customContent, t, lang)).ServeHTTP(w, r)
			},
		)
		r.Get("/",
			func(w http.ResponseWriter, r *http.Request) {
				templ.Handler(views.PostRedirect()).ServeHTTP(w, r)
			},
		)
	})
	r.Route("/language", func(r chi.Router) {
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

				expiryDuration := 1 * time.Hour * 24 * 30
				expiresAt := time.Now().Add(expiryDuration)

				language := data["language"][0]
				cookie := http.Cookie{
					Name:  "Language",
					Value: language,
					Path:  "/",
					// sets the expiry time also used in the session
					Expires:  expiresAt,
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
				}

				http.SetCookie(w, &cookie)

				// refreshes everything so the language preference gets saved
				w.Header().Set("HX-Refresh", "true")

				w.WriteHeader(http.StatusNoContent)
			},
		)
	})
	r.Route("/search", func(r chi.Router) {
		r.Get("/",
			func(w http.ResponseWriter, r *http.Request) {
				lang := translations.GetLanguageFromCookie(r)
				t := translations.GetTranslations(lang)
				isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
				templ.Handler(views.SearchPage(customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
			},
		)
		r.Get("/{searchTerm}/{index}",
			func(w http.ResponseWriter, r *http.Request) {
				isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
				searchTerm := chi.URLParam(r, "searchTerm")
				index, err := strconv.ParseUint(chi.URLParam(r, "index"), 10, 0)
				if err != nil {
					fmt.Println("failed parsing")
					lang := translations.GetLanguageFromCookie(r)
					t := translations.GetTranslations(lang)
					templ.Handler(views.GenericError(t.InvalidSearchRange, customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
				}
				posts, lastPage := controllers.SearchPostList(env, searchTerm, uint(index))
				lang := translations.GetLanguageFromCookie(r)
				t := translations.GetTranslations(lang)
				templ.Handler(views.SearchResults(posts, index, lastPage, customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
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

				searchTerm := data["searchTerm"][0]
				templ.Handler(views.RedirectTo("search/"+searchTerm+"/0")).ServeHTTP(w, r)
			},
		)
	})
	r.Get("/attachments/{postID}/{fileName}",
		func(w http.ResponseWriter, r *http.Request) {
			postID, err := strconv.ParseUint(chi.URLParam(r, "postID"), 10, 0)
			if err != nil {
				lang := translations.GetLanguageFromCookie(r)
				t := translations.GetTranslations(lang)
				isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
				templ.Handler(views.GenericError(t.InvalidAttachmentLocation+fmt.Sprint(postID), customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
			}
			fileName := chi.URLParam(r, "fileName")

			var attachment models.Attachment
			env.DB.Where(&models.Attachment{PostID: uint(postID), Filename: fileName}).First(&attachment)

			http.ServeFile(w, r, "./attachments/"+fmt.Sprint(postID)+"/"+fileName)
		})
	r.Route("/post", func(r chi.Router) {
		r.Post("/",
			func(w http.ResponseWriter, r *http.Request) {
				err := r.ParseMultipartForm(10 << 20) // Limit file upload size to 10MB
				if err != nil {
					t := translations.GetTranslations(translations.GetLanguageFromCookie(r))
					http.Error(w, t.FailedToParseFormData, http.StatusBadRequest)
					return
				}

				isLoggedIn, session := controllers.GetLoginFromSession(env, r)
				if isLoggedIn {
					currentTime := time.Now()
					post := &models.RootPost{
						Title:        r.FormValue("title"),
						Body:         r.FormValue("body"),
						CreationDate: currentTime,
						UpdateDate:   currentTime,
						Op:           session.UserAccount.Name,
					}

					result := env.DB.Create(&post)
					if result.Error != nil {
						fmt.Println("failed to create post")
					}

					files := r.MultipartForm.File["attachments"]
					for _, fileHeader := range files {
						file, err := fileHeader.Open()
						if err != nil {
							env.DB.Delete(&post)
							t := translations.GetTranslations(translations.GetLanguageFromCookie(r))
							templ.Handler(views.Message(t.FailedToAddAttachments)).ServeHTTP(w, r)
							return
						}
						defer file.Close()

						attachmentPath := "./attachments"

						if _, err := os.Stat(attachmentPath + "/" + fmt.Sprint(post.ID)); os.IsNotExist(err) {
							err := os.Mkdir(attachmentPath+"/"+fmt.Sprint(post.ID), 0755)
							if err != nil {
								env.DB.Delete(&post)
								t := translations.GetTranslations(translations.GetLanguageFromCookie(r))
								templ.Handler(views.Message(t.FailedToAddAttachments)).ServeHTTP(w, r)
								return
							}
						}

						filePath := attachmentPath + "/" + fmt.Sprint(post.ID) + "/" + fileHeader.Filename

						outFile, err := os.Create(filePath)
						if err != nil {
							env.DB.Delete(&post)
							t := translations.GetTranslations(translations.GetLanguageFromCookie(r))
							templ.Handler(views.Message(t.FailedToAddAttachments)).ServeHTTP(w, r)
							return
						}
						defer outFile.Close()

						_, err = io.Copy(outFile, file)
						if err != nil {
							env.DB.Delete(&post)
							t := translations.GetTranslations(translations.GetLanguageFromCookie(r))
							templ.Handler(views.Message(t.FailedToAddAttachments)).ServeHTTP(w, r)
							return
						}

						// Create the attachment record and associate it with the post
						attachment := models.Attachment{
							PostID:   post.ID,
							Filename: fileHeader.Filename,
						}
						env.DB.Create(&attachment)
					}
					// gets a new session token
					controllers.RefreshSession(env, w, r)

					t := translations.GetTranslations(translations.GetLanguageFromCookie(r))
					w.Header().Set("HX-Refresh", "true")
					templ.Handler(views.Message(t.PostedSucessfully)).ServeHTTP(w, r)
				}
			},
		)
		r.Post("/preview",
			func(w http.ResponseWriter, r *http.Request) {
				err := r.ParseMultipartForm(10 << 20) // Limit file upload size to 10MB
				if err != nil {
					t := translations.GetTranslations(translations.GetLanguageFromCookie(r))
					http.Error(w, t.FailedToParseFormData, http.StatusBadRequest)
					return
				}

				input := r.FormValue("body")

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
					lang := translations.GetLanguageFromCookie(r)
					t := translations.GetTranslations(lang)
					isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
					templ.Handler(views.GenericError(t.InvalidPostID, customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
				}
				templ.Handler(views.RedirectTo("post/"+fmt.Sprint(postId)+"/0")).ServeHTTP(w, r)
			},
		)
		r.Get("/{postId}/{commentIndex}",
			func(w http.ResponseWriter, r *http.Request) {
				isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
				// 10 is base 10 and 0 indicates parsing into system-size int
				postId, err := strconv.ParseUint(chi.URLParam(r, "postId"), 10, 0)
				if err != nil {
					lang := translations.GetLanguageFromCookie(r)
					t := translations.GetTranslations(lang)
					templ.Handler(views.GenericError(t.InvalidPostID, customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
				}

				var post models.RootPost
				result := env.DB.First(&post, postId)
				lang := translations.GetLanguageFromCookie(r)
				t := translations.GetTranslations(lang)
				if result.Error != nil {
					templ.Handler(views.GenericError(t.FailedToLoadPosts, customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
				} else {
					comments := controllers.GetCommentList(env, uint(postId))
					attachments := controllers.GetAttachmentList(env, uint(postId))
					isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
					templ.Handler(views.Post(post, comments, attachments, customContent, isLoggedIn, t, lang)).ServeHTTP(w, r)
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
				lang := translations.GetLanguageFromCookie(r)
				t := translations.GetTranslations(lang)

				data, err := url.ParseQuery(response)
				if err != nil {
					fmt.Println("Failed to parse query")
				}

				isLoggedIn, session := controllers.GetLoginFromSession(env, r)
				if isLoggedIn {
					currentTime := time.Now()

					postId, err := strconv.ParseUint(data["rootPostID"][0], 10, 0)
					if err != nil {
						templ.Handler(views.GenericError(t.InvalidPostID, customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
					}

					fmt.Println("Creating a comment with root post id \"" + fmt.Sprint(postId) + "\" and body \"" + data["text"][0])
					env.DB.Create(&models.Comment{
						RootPostID:   uint(postId),
						Body:         data["text"][0],
						CreationDate: currentTime,
						UpdateDate:   currentTime,
						Op:           session.UserAccount.Name,
					})
					// gets a new session token
					controllers.RefreshSession(env, w, r)
					w.Header().Set("HX-Refresh", "true")
					w.WriteHeader(http.StatusNoContent)
				}
			},
		)
	})
	r.Route("/register", func(r chi.Router) {
		r.Get("/",
			func(w http.ResponseWriter, r *http.Request) {
				isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
				lang := translations.GetLanguageFromCookie(r)
				t := translations.GetTranslations(lang)
				templ.Handler(views.Register(customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
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
					lang := translations.GetLanguageFromCookie(r)
					t := translations.GetTranslations(lang)
					isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
					templ.Handler(views.GenericError(t.UsernameAlreadyTaken, customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
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
				lang := translations.GetLanguageFromCookie(r)
				t := translations.GetTranslations(lang)
				isLoggedIn, _ := controllers.GetLoginFromSession(env, r)
				templ.Handler(views.Login(redirect, customContent, t, lang, isLoggedIn)).ServeHTTP(w, r)
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
				lang := translations.GetLanguageFromCookie(r)
				t := translations.GetTranslations(lang)
				templ.Handler(views.Logout(isLoggedIn, customContent, t, lang)).ServeHTTP(w, r)
			},
		)
		r.Post("/",
			func(w http.ResponseWriter, r *http.Request) {
				controllers.RemoveSession(env, w, r)
				templ.Handler(views.RedirectTo("posts")).ServeHTTP(w, r)
			},
		)
	})
	r.Route("/sessions", func(r chi.Router) {
		r.Get("/",
			func(w http.ResponseWriter, r *http.Request) {
				isLoggedIn, session := controllers.GetLoginFromSession(env, r)

				sessionList := controllers.GetSessionsForUser(env, r, session)

				lang := translations.GetLanguageFromCookie(r)
				t := translations.GetTranslations(lang)
				templ.Handler(views.SessionList(isLoggedIn, sessionList, customContent, t, lang)).ServeHTTP(w, r)
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
