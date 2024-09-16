package controllers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jkulzer/foryoum/v2/db"
	"github.com/jkulzer/foryoum/v2/models"
	"net/http"
	"time"
)

func IsExpired(s models.Session) bool {
	return s.Expiry.Before(time.Now())
}

func NewSession(env *db.Env, userAccount models.UserAccount) (string, time.Duration) {
	sessionToken := uuid.NewString()
	expiryDuration := 120 * time.Second
	// 120 seconds expiry time
	expiresAt := time.Now().Add(expiryDuration)

	env.DB.Create(&models.Session{
		Token:       sessionToken,
		UserAccount: userAccount,
		Expiry:      expiresAt,
	})

	fmt.Println(userAccount.Name)

	return sessionToken, expiryDuration
}

func GetLoginFromSession(env *db.Env, r *http.Request) (bool, models.Session) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		fmt.Println("Failed to read cookie")
		return false, models.Session{} // returns empty UserAccount struct
	}

	var session models.Session

	// checks if the token in the cookie is in any active session
	result := env.DB.Where(&models.Session{Token: cookie.Value}).First(&session)
	if result.Error != nil {
		return false, models.Session{} // returns empty UserAccount struct
	} else {
		return true, session
	}

}
