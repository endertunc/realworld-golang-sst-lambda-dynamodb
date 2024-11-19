package generator

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"log"
	"os"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"strconv"
)

func init() {
	seed, ok := os.LookupEnv("GOFAKEIT_SEED")
	if ok {
		seedInt, err := strconv.ParseUint(seed, 10, 64)
		if err != nil {
			log.Fatalf("GOFAKEIT_SEED must be a valid intiger: %v", err)
		}
		gofakeit.GlobalFaker = gofakeit.New(seedInt)
	}
}

func GenerateArticle() domain.Article {
	title := gofakeit.LoremIpsumSentence(gofakeit.Number(5, 10))
	date := gofakeit.PastDate()
	return domain.Article{
		Id:             generateUUID(),
		Title:          title,
		Slug:           slug.Make(title),
		Description:    gofakeit.LoremIpsumSentence(gofakeit.Number(10, 20)),
		Body:           gofakeit.LoremIpsumParagraph(2, 20, 100, "\n"),
		TagList:        []string{gofakeit.LoremIpsumWord(), gofakeit.LoremIpsumWord()},
		FavoritesCount: gofakeit.Number(0, 100),
		AuthorId:       generateUUID(),
		CreatedAt:      date,
		UpdatedAt:      date,
	}
}

func GenerateComment() domain.Comment {
	date := gofakeit.PastDate()
	return domain.Comment{
		Id:        generateUUID(),
		ArticleId: generateUUID(),
		AuthorId:  generateUUID(),
		Body:      gofakeit.LoremIpsumSentence(gofakeit.Number(10, 50)),
		CreatedAt: date,
		UpdatedAt: date,
	}
}

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
	return domain.User{
		Id:       generateUUID(),
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Bio:      bio,
		Image:    image,
	}
}

func generateUUID() uuid.UUID {
	return uuid.MustParse(gofakeit.UUID())
}
