package main

import (
	"fmt"
	"net/http"
	"os"
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

	customContent, err := os.ReadFile("./custom.html")
	if err != nil {
		fmt.Println("No custom content detected")
	}
	mainPage, err := os.ReadFile("./mainPage.html")
	if err != nil {
		fmt.Println("No main page set up! This will confuse users. The main page can be created in \"mainPage.html\"")
	}

	attachmentPath := "./attachments"

	if _, err := os.Stat(attachmentPath); os.IsNotExist(err) {
		err := os.Mkdir(attachmentPath, 0755)
		if err != nil {
			panic(err)
		}
		fmt.Println("Attachment Folder created successfully")
	} else {
		fmt.Println("Attachment Folder already exists")
	}

	routes.Router(r, env, string(customContent), string(mainPage))

	err = http.ListenAndServe(":"+strconv.Itoa(port), r)
	if err != nil {
		panic(err)
	}

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
