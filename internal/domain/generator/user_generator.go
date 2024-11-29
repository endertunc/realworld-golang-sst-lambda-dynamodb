package generator

import (
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

func GenerateUser() domain.User {
	var bio *string
	if gofakeit.Bool() {
		quote := gofakeit.Quote()
		bio = &quote
	}
	var image *string
	if gofakeit.Bool() {
		url := gofakeit.URL()
		image = &url
	}
	date := gofakeit.PastDate().Truncate(time.Millisecond)
	return domain.User{
		Id:             uuid.New(),
		Username:       gofakeit.Username(),
		Email:          gofakeit.Email(),
		Bio:            bio,
		Image:          image,
		HashedPassword: gofakeit.LetterN(64),
		CreatedAt:      date,
		UpdatedAt:      date,
	}
}
