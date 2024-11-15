package generator

import (
	"github.com/brianvoe/gofakeit/v7"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

func GenerateAddCommentRequestBodyDTO() dto.AddCommentRequestBodyDTO {
	return dto.AddCommentRequestBodyDTO{
		Comment: GenerateAddCommentRequestDTO(),
	}
}

func GenerateAddCommentRequestDTO() dto.AddCommentRequestDTO {
	return dto.AddCommentRequestDTO{
		Body: gofakeit.LoremIpsumSentence(gofakeit.Number(10, 30)),
	}
}
