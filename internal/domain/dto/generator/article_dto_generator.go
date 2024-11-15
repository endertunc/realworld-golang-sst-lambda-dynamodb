package generator

import (
	"github.com/brianvoe/gofakeit/v7"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

func GenerateCreateArticleRequestBodyDTO() dto.CreateArticleRequestBodyDTO {
	return dto.CreateArticleRequestBodyDTO{
		Article: GenerateCreateArticleRequestDTO(),
	}
}

func GenerateCreateArticleRequestDTO() dto.CreateArticleRequestDTO {
	return dto.CreateArticleRequestDTO{
		Title:       gofakeit.LoremIpsumSentence(gofakeit.Number(5, 10)),
		Description: gofakeit.LoremIpsumSentence(gofakeit.Number(5, 15)),
		Body:        gofakeit.LoremIpsumParagraph(1, 5, 10, "\n"),
		TagList:     []string{gofakeit.LoremIpsumWord(), gofakeit.LoremIpsumWord()},
	}
}
