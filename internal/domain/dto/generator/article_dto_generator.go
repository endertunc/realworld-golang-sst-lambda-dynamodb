package generator

import (
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"

	"github.com/brianvoe/gofakeit/v7"
)

//func GenerateCreateArticleRequestBodyDTO() dto.CreateArticleRequestBodyDTO {
//	return dto.CreateArticleRequestBodyDTO{
//		Article: GenerateCreateArticleRequestDTO(),
//	}
//}

func GenerateCreateArticleRequestDTO() dto.CreateArticleRequestDTO {
	return dto.CreateArticleRequestDTO{
		Title:       gofakeit.LoremIpsumSentence(gofakeit.Number(5, 10)),
		Description: gofakeit.LoremIpsumSentence(gofakeit.Number(5, 15)),
		Body:        gofakeit.LoremIpsumParagraph(1, 5, 10, "\n"),
		TagList:     []string{gofakeit.LoremIpsumWord(), gofakeit.LoremIpsumWord()},
	}
}

func GenerateUpdateArticleRequestDTO() dto.UpdateArticleRequestDTO {
	title := gofakeit.LoremIpsumSentence(gofakeit.Number(5, 10))
	description := gofakeit.LoremIpsumSentence(gofakeit.Number(5, 15))
	body := gofakeit.LoremIpsumParagraph(1, 5, 10, "\n")

	return dto.UpdateArticleRequestDTO{
		Title:       &title,
		Description: &description,
		Body:        &body,
	}
}
