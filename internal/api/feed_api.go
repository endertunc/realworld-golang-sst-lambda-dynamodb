package api

import (
	"context"
	"github.com/google/uuid"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
)

type UserFeedApi struct {
	feedService      service.FeedServiceInterface
	paginationConfig PaginationConfig
}

func NewUserFeedApi(feedService service.FeedServiceInterface, paginationConfig PaginationConfig) UserFeedApi {
	return UserFeedApi{
		feedService:      feedService,
		paginationConfig: paginationConfig,
	}
}

func (uf UserFeedApi) FetchUserFeed(ctx context.Context, w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	limit, ok := GetIntQueryParamOrDefault(ctx, w, r, "limit", uf.paginationConfig.DefaultLimit, &uf.paginationConfig.MinLimit, &uf.paginationConfig.MaxLimit)
	if !ok {
		return
	}

	nextPageToken, ok := GetOptionalStringQueryParam(w, r, "offset")
	if !ok {
		return
	}

	feedItems, nextToken, err := uf.feedService.FetchArticlesFromFeed(ctx, userId, limit, nextPageToken)
	if err != nil {
		ToInternalServerHTTPError(w, err)
		return
	}
	resp := dto.ToMultipleArticlesResponseBodyDTO(feedItems, nextToken)
	ToSuccessHTTPResponse(w, resp)
	return
}
