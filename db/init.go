package db

import (
	"github.com/jkulzer/foryoum/v2/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Env struct {
	DB *gorm.DB
}

func Init() *Env {
	db, err := gorm.Open(
		sqlite.Open("sqlite.db"),
		&gorm.Config{
			TranslateError: true,
		},
	)
	if err != nil {
		panic("failed to connect database")
	}
	// create tables for all structs
	db.AutoMigrate(&models.RootPost{})
	db.AutoMigrate(&models.Comment{})
	db.AutoMigrate(&models.UserAccount{})
	db.AutoMigrate(&models.Session{})

	env := &Env{DB: db}

	return env
}
