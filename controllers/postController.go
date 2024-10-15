package controllers

import (
	"github.com/jkulzer/foryoum/v2/db"
	"github.com/jkulzer/foryoum/v2/models"
	"strings"
)

func GetPostList(env *db.Env, listIndex uint) ([]models.RootPost, bool) {
	var posts []models.RootPost
	pageLength := uint(10)
	offset := listIndex * pageLength

	// Query the database with limit, offset, and order by creation time (newest first)
	queryResult := env.DB.Order("created_at DESC").Limit(int(pageLength)).Offset(int(offset)).Find(&posts)
	if queryResult.Error != nil {
		return posts, true
	}

	// Check if no posts were found
	if len(posts) < int(pageLength) {
		return posts, true
	}

	return posts, false
}

func SearchPostList(env *db.Env, searchTerm string, listIndex uint) ([]models.RootPost, bool) {
	var posts []models.RootPost
	pageLength := uint(10)
	offset := listIndex * pageLength

	searchTerm = strings.TrimSpace(searchTerm)

	queryResult := env.DB.Where("title LIKE ? OR body LIKE ?", "%"+searchTerm+"%", "%"+searchTerm+"%").
		Limit(int(pageLength)).
		Offset(int(offset)).
		Find(&posts)

	if queryResult.Error != nil {
		return posts, true
	}

	// only returns false if there is pagination needed
	if len(posts) == 0 {
		return posts, true
	}

	return posts, false
}

func GetCommentList(env *db.Env, postID uint) []models.Comment {
	var comments []models.Comment
	env.DB.Where("root_post_id = ?", postID).Find(&comments)
	return comments
}

func GetAttachmentList(env *db.Env, postID uint) []models.Attachment {
	var attachments []models.Attachment
	env.DB.Where(&models.Attachment{PostID: postID}).First(&attachments)
	return attachments
}
