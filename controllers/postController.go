package controllers

import (
	"github.com/jkulzer/foryoum/v2/db"
	"github.com/jkulzer/foryoum/v2/models"
)

func GetPostList(length uint, env *db.Env, offset uint) []models.RootPost {
	var posts []models.RootPost
	for index := 1 + offset; index <= length; index++ {
		var newPost models.RootPost
		queryResult := env.DB.First(&newPost, index)
		if queryResult.Error != nil {
			break
		}
		posts = append(posts, newPost)
	}

	return posts
}
