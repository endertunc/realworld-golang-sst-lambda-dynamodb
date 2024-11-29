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

func init() {
	// ToDo @ender I think this is a good way to ensure uniqueness. this will give away the creation time of the article tho.
	slug.AppendTimestamp = true
}

func NewArticle(title, description, body string, tagList []string, authorId uuid.UUID) Article {
	now := time.Now().Truncate(time.Millisecond)
	return Article{
		Id:             uuid.New(),
		Title:          title,
		Slug:           GenerateSlug(title),
		Description:    description,
		Body:           body,
		TagList:        tagList,
		FavoritesCount: 0,
		AuthorId:       authorId,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func GenerateSlug(title string) string {
	return slug.Make(title)
}
