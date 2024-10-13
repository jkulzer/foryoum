package controllers

import (
	"github.com/jkulzer/foryoum/v2/db"
	"github.com/jkulzer/foryoum/v2/models"
	"strings"
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

func SearchPostList(env *db.Env, searchTerm string) ([]models.RootPost, bool) {
	var posts []models.RootPost
	searchTerm = strings.TrimSpace(searchTerm)
	queryResult := env.DB.Where("title LIKE ? OR body LIKE ?", "%"+searchTerm+"%", "%"+searchTerm+"%").Find(&posts)
	if queryResult.Error != nil {
		return posts, true
	}
	return posts, false
}

func GetCommentList(env *db.Env, postID uint) []models.Comment {
	var comments []models.Comment
	env.DB.Where("root_post_id = ?", postID).Find(&comments)
	return comments
}
