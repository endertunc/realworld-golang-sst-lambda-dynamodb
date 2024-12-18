package api

import (
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/service"

	"github.com/google/uuid"
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

func (uf UserFeedApi) FetchUserFeed(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	ctx := r.Context()

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
}
