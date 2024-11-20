package domain

import (
	"github.com/google/uuid"
	"time"
)

type Comment struct {
	Id        uuid.UUID
	ArticleId uuid.UUID
	AuthorId  uuid.UUID
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewComment(articleId, authorId uuid.UUID, body string) Comment {
	now := time.Now().Truncate(time.Millisecond)
	return Comment{
		Id:        uuid.New(),
		ArticleId: articleId,
		AuthorId:  authorId,
		Body:      body,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
