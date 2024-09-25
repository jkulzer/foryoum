package routes

import (
	"github.com/jkulzer/foryoum/v2/controllers"
	"github.com/jkulzer/foryoum/v2/db"
	"net/http"
)

func RefreshToken(env *db.Env) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			controllers.RefreshSession(env, w, r)
			next.ServeHTTP(w, r)
		})
	}
}
