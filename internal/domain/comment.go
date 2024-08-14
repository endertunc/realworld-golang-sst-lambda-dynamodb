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
