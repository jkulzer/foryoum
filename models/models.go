package models

import (
	"github.com/google/uuid"
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
	Attachments  []Attachment `gorm:"foreignKey:PostID"`
	Op           string
}

type Comment struct {
	gorm.Model
	RootPostID   uint
	RootPost     RootPost
	Body         string
	CreationDate time.Time
	UpdateDate   time.Time
	Op           string
}

type Attachment struct {
	ID       uint `gorm:"primaryKey"`
	PostID   uint `gorm:"index"` // Foreign key to the Post
	Filename string
}

type UserAccount struct {
	gorm.Model
	Name     string `gorm:"unique"`
	Password string
	Sessions []Session
}

type Session struct {
	gorm.Model
	Token         uuid.UUID
	UserAccountID uint
	UserAccount   UserAccount
	Expiry        time.Time
}
