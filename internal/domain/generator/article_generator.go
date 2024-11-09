package generator

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
)

// Uses math/rand/v2(PCG Pseudo) with mutex locking
var faker = gofakeit.GlobalFaker

func GenerateArticle() domain.Article {
	title := faker.LoremIpsumSentence(faker.Number(5, 10))
	date := faker.PastDate()
	return domain.Article{
		Id:             uuid.MustParse(faker.UUID()),
		Title:          title,
		Slug:           slug.Make(title),
		Description:    faker.LoremIpsumSentence(faker.Number(10, 20)),
		Body:           faker.LoremIpsumParagraph(2, 20, 100, "\n"),
		TagList:        []string{faker.LoremIpsumWord(), faker.LoremIpsumWord()},
		FavoritesCount: faker.Number(0, 100),
		AuthorId:       uuid.MustParse(faker.UUID()),
		CreatedAt:      date,
		UpdatedAt:      date,
	}
}

func GenerateComment() domain.Comment {
	date := faker.PastDate()
	return domain.Comment{
		Id:        uuid.MustParse(faker.UUID()),
		Body:      faker.LoremIpsumSentence(faker.Number(10, 50)),
		AuthorId:  uuid.MustParse(faker.UUID()),
		CreatedAt: date,
		UpdatedAt: date,
	}
}

func GenerateUser() domain.User {
	var bio *string
	if faker.Bool() {
		quote := faker.Quote()
		bio = &quote
	}
	var image *string
	if faker.Bool() {
		url := faker.URL()
		image = &url
	}
	return domain.User{
		Id:       uuid.MustParse(faker.UUID()),
		Username: faker.Username(),
		Email:    faker.Email(),
		Bio:      bio,
		Image:    image,
	}
}
