package domain

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	Id             uuid.UUID
	Email          string
	HashedPassword string
	Username       string
	Bio            *string
	Image          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewUser(email, username, hashedPassword string) User {
	now := time.Now().Truncate(time.Millisecond)
	return User{
		Id:             uuid.New(),
		Email:          email,
		HashedPassword: hashedPassword,
		Username:       username,
		Bio:            nil,
		Image:          nil,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

}
