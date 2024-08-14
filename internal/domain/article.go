package domain

import (
	"github.com/google/uuid"
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
