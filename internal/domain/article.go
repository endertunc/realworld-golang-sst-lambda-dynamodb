package domain

import (
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"time"
)

type Article struct {
	Id             uuid.UUID
	Title          string
	Slug           string
	Description    string
	Body           string
	TagList        []string
	FavoritesCount int
	AuthorId       uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewArticle(title, description, body string, tagList []string, authorId uuid.UUID) Article {
	now := time.Now()
	return Article{
		Id:             uuid.New(),
		Title:          title,
		Slug:           slug.Make(title),
		Description:    description,
		Body:           body,
		TagList:        tagList,
		FavoritesCount: 0,
		AuthorId:       authorId,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
