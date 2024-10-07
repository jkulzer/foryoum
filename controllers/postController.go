package controllers

import (
	"github.com/jkulzer/foryoum/v2/db"
	"github.com/jkulzer/foryoum/v2/models"
)

func GetPostList(env *db.Env, listIndex uint) ([]models.RootPost, bool) {
	var posts []models.RootPost
	pageLength := uint(10)
	beginning := 1 + listIndex*pageLength
	for index := beginning; index <= beginning+pageLength; index++ {
		var newPost models.RootPost
		queryResult := env.DB.First(&newPost, index)
		if queryResult.Error != nil {
			return posts, true
		}
		posts = append(posts, newPost)
	}
	return posts, false
}
