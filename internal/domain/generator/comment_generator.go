package generator

import (
	"realworld-aws-lambda-dynamodb-golang/internal/domain"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

func GenerateComment() domain.Comment {
	date := gofakeit.PastDate()
	return domain.Comment{
		Id:        uuid.New(),
		ArticleId: uuid.New(),
		AuthorId:  uuid.New(),
		Body:      gofakeit.LoremIpsumSentence(gofakeit.Number(10, 50)),
		CreatedAt: date,
		UpdatedAt: date,
	}
}

func GenerateCommentWithArticleId(articleId uuid.UUID) domain.Comment {
	comment := GenerateComment()
	comment.ArticleId = articleId
	return comment
}
