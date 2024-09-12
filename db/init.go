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
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&models.RootPost{})

	env := &Env{DB: db}

	return env
}
