package models

import (
	"gorm.io/gorm"
	"time"
)

type RootPost struct {
	gorm.Model
	Title        string
	Body         string
	CreationDate time.Time
	UpdateDate   time.Time
	Op           string
	Version      int
}
