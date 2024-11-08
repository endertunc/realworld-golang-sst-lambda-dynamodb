package user

import (
	"context"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"time"
)

type FeedApi struct {
	FeedService FeedServiceInterface
}

func NewFeedApi(feedService FeedServiceInterface) FeedApi {
	return FeedApi{
		FeedService: feedService,
	}
}

type FeedServiceInterface interface {
	FanoutArticle(ctx context.Context, articleId, authorId uuid.UUID, createdAt time.Time) error
	FetchArticlesFromFeed(ctx context.Context, userId uuid.UUID, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error)
}

func (uf FeedApi) FetchUserFeed(ctx context.Context, userId uuid.UUID, limit int, nextPageToken *string) (dto.MultipleArticlesResponseBodyDTO, error) {
	feedItems, nextToken, err := uf.FeedService.FetchArticlesFromFeed(ctx, userId, limit, nextPageToken)
	if err != nil {
		return dto.MultipleArticlesResponseBodyDTO{}, err
	}
	return dto.ToMultipleArticlesResponseBodyDTO(feedItems, nextToken), nil
}
