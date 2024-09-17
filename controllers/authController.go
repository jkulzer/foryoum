package controllers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jkulzer/foryoum/v2/db"
	"github.com/jkulzer/foryoum/v2/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

func IsExpired(s models.Session) bool {
	return s.Expiry.Before(time.Now())
}

func NewSession(env *db.Env, userAccount models.UserAccount) (string, time.Duration) {
	sessionToken := uuid.NewString()
	// 5 min expiry time
	expiryDuration := 5 * time.Minute
	expiresAt := time.Now().Add(expiryDuration)

	env.DB.Create(&models.Session{
		Token:         sessionToken,
		UserAccountID: userAccount.ID,
		Expiry:        expiresAt,
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

	env.DB.Preload("UserAccount").Where(&models.Session{Token: cookie.Value}).First(&session)
	// checks if the token in the cookie is in any active session
	result := env.DB.Where(&models.Session{Token: cookie.Value}).First(&session)
	if result.Error != nil {
		return false, models.Session{} // returns empty UserAccount struct
	} else {
		return true, session
	}

}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateSession(env *db.Env, userAccount models.UserAccount, w http.ResponseWriter) {
	sessionToken, expiryDuration := NewSession(env, userAccount)
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
	fmt.Println("New Session for \"" + userAccount.Name + "\"")

	http.SetCookie(w, &cookie)

}

func RefreshSession(env *db.Env, w http.ResponseWriter, r *http.Request) {
	user := RemoveSession(env, w, r)
	CreateSession(env, user, w)
}

func RemoveSession(env *db.Env, w http.ResponseWriter, r *http.Request) models.UserAccount {
	_, session := GetLoginFromSession(env, r)

	user := session.UserAccount

	env.DB.Delete(&session)

	// deletes the cookie
	cookie := http.Cookie{
		Name:     "Session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	return user
}

func GetSessionsForUser(env *db.Env, r *http.Request, session models.Session) []models.Session {

	var sessionList []models.Session
	result := env.DB.Find(&sessionList).Where(models.Session{UserAccountID: session.UserAccountID})
	if result.Error != nil {
		fmt.Println("Failed to get all user sessions for user " + session.UserAccount.Name)
	}

	return sessionList
}
