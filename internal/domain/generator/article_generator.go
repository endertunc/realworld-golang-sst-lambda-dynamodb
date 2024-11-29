package generator

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
)

func GenerateArticle() domain.Article {
	title := gofakeit.LoremIpsumSentence(gofakeit.Number(5, 10))
	date := gofakeit.PastDate()
	return domain.Article{
		Id:             uuid.New(),
		Title:          title,
		Slug:           slug.Make(title),
		Description:    gofakeit.LoremIpsumSentence(gofakeit.Number(10, 20)),
		Body:           gofakeit.LoremIpsumParagraph(2, 20, 100, "\n"),
		TagList:        []string{gofakeit.LoremIpsumWord(), gofakeit.LoremIpsumWord()},
		FavoritesCount: gofakeit.Number(0, 100),
		AuthorId:       uuid.New(),
		CreatedAt:      date,
		UpdatedAt:      date,
	}
}
