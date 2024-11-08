package api

import (
	"context"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
)

type FeedApi struct {
	FeedService service.FeedServiceInterface
}

func NewFeedApi(feedService service.FeedServiceInterface) FeedApi {
	return FeedApi{
		FeedService: feedService,
	}
}

func (uf FeedApi) FetchUserFeed(ctx context.Context, userId uuid.UUID, limit int, nextPageToken *string) (dto.MultipleArticlesResponseBodyDTO, error) {
	feedItems, nextToken, err := uf.FeedService.FetchArticlesFromFeed(ctx, userId, limit, nextPageToken)
	if err != nil {
		return dto.MultipleArticlesResponseBodyDTO{}, err
	}
	return dto.ToMultipleArticlesResponseBodyDTO(feedItems, nextToken), nil
}
