package models

import (
	"gorm.io/gorm"
	"time"
)

// MAKE SURE TO KEEP THE LIST OF MIGRATIONS IN `db/init.go` UP TO DATE!!!!!
//WARNING !!!!!!!

type RootPost struct {
	gorm.Model
	Title        string
	Body         string
	CreationDate time.Time
	UpdateDate   time.Time
	Op           string
	Version      int
}

type UserAccount struct {
	gorm.Model
	Name     string `gorm:"unique"`
	Password string
	Sessions []Session
}

type Session struct {
	gorm.Model
	Token         string
	UserAccountID uint
	UserAccount   UserAccount
	Expiry        time.Time
}
