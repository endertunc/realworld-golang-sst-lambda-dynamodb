package api

import (
	"context"
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
)

type CommentApi struct {
	commentService service.CommentServiceInterface
	userService    service.UserServiceInterface
	profileService service.ProfileServiceInterface
}

func NewCommentApi(commentService service.CommentServiceInterface, userService service.UserServiceInterface, profileService service.ProfileServiceInterface) CommentApi {
	return CommentApi{
		commentService: commentService,
		userService:    userService,
		profileService: profileService,
	}
}

// GetArticleComments
/**
 * In a "realworld" application, this function could be even optimized further
 * by using a single query to get all authors and their "isFollowing" information using BatchGetItem.
 * Following two queries can be combined into one BatchGetItem query:
 * - fetch users by their ids (it will be used as comment author )
 * - fetch "isFollowing" status for each user (to populate author.following field)
 *
 * However, this complexity might not be needed with an alternative approach such as caching
 * with DAX (DynamoDB Accelerator) or other caching mechanisms.
 *
 * Regardless, one must monitor the performance of the application and optimize accordingly.
 */
func (aa CommentApi) GetArticleComments(ctx context.Context, w http.ResponseWriter, r *http.Request, loggedInUserId *uuid.UUID) {
	slug, ok := GetPathParamHTTP(ctx, w, r, "slug")
	if !ok {
		return
	}

	handleError := func(err error) {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusNotFound, "article not found")
			return
		}
		ToInternalServerHTTPError(w, err)
	}

	comments, err := aa.commentService.GetArticleComments(ctx, slug)
	if err != nil {
		handleError(err)
		return
	}
	if len(comments) == 0 {
		resp := dto.MultiCommentsResponseBodyDTO{Comment: []dto.CommentResponseDTO{}}
		ToSuccessHTTPResponse(w, resp)
		return
	} else {
		authorIdsMap := make(map[uuid.UUID]struct{})
		for _, comment := range comments {
			authorIdsMap[comment.AuthorId] = struct{}{}
		}

		uniqueAuthorIdsList := make([]uuid.UUID, 0, len(authorIdsMap))
		for k := range authorIdsMap {
			uniqueAuthorIdsList = append(uniqueAuthorIdsList, k)
		}

		authors, err := aa.userService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
		if err != nil {
			handleError(err)
			return
		}

		authorIdsToAuthorMap := make(map[uuid.UUID]domain.User, len(authors))
		for _, author := range authors {
			authorIdsToAuthorMap[author.Id] = author
		}

		if loggedInUserId == nil {
			resp := dto.ToMultiCommentsResponseBodyDTO(comments, authorIdsToAuthorMap, mapset.NewSetWithSize[uuid.UUID](0))
			ToSuccessHTTPResponse(w, resp)
			return
		} else {
			followedAuthorsSet, err := aa.profileService.IsFollowingBulk(ctx, *loggedInUserId, uniqueAuthorIdsList)
			if err != nil {
				handleError(err)
				return
			}
			resp := dto.ToMultiCommentsResponseBodyDTO(comments, authorIdsToAuthorMap, followedAuthorsSet)
			ToSuccessHTTPResponse(w, resp)
			return
		}
	}
}

func (aa CommentApi) AddComment(ctx context.Context, w http.ResponseWriter, r *http.Request, loggedInUserId uuid.UUID) {
	// Get slug from the request path
	slug, ok := GetPathParamHTTP(ctx, w, r, "slug")
	if !ok {
		return
	}

	// Parse request body
	addCommentRequestBodyDTO, ok := ParseAndValidateBody[dto.AddCommentRequestBodyDTO](ctx, w, r)
	if !ok {
		return
	}

	handleError := func(err error) {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusNotFound, "article not found")
			return
		}
		ToInternalServerHTTPError(w, err)
	}

	comment, err := aa.commentService.AddComment(ctx, loggedInUserId, slug, addCommentRequestBodyDTO.Comment.Body)
	if err != nil {
		handleError(err)
		return
	}

	user, err := aa.userService.GetUserByUserId(ctx, loggedInUserId)
	if err != nil {
		handleError(err)
		return
	}

	// the current user is the author, and the user can't follow itself,
	// thus we simply pass isFollowing as false
	resp := dto.ToSingleCommentResponseBodyDTO(comment, user, false)

	// Success response
	ToSuccessHTTPResponse(w, resp)
}

func (aa CommentApi) DeleteComment(ctx context.Context, w http.ResponseWriter, r *http.Request, loggedInUserId uuid.UUID) {
	slug, ok := GetPathParamHTTP(ctx, w, r, "slug")
	if !ok {
		return
	}

	commentIdAsString, ok := GetPathParamHTTP(ctx, w, r, "id")
	if !ok {
		return
	}

	commentId, err := uuid.Parse(commentIdAsString)
	if err != nil {
		slog.DebugContext(ctx, "invalid commentId path param", slog.String("commentId", commentIdAsString), slog.Any("error", err))
		ToSimpleHTTPError(w, http.StatusBadRequest, "commentId path parameter must be a valid UUID")
		return
	}

	err = aa.commentService.DeleteComment(ctx, loggedInUserId, slug, commentId)
	if err != nil {
		if errors.Is(err, errutil.ErrCommentNotFound) {
			slog.DebugContext(ctx, "comment not found", slog.String("slug", slug), slog.String("commentId", commentId.String()), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusNotFound, "comment not found")
			return
		} else if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug), slog.String("commentId", commentId.String()), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusNotFound, "article not found")
			return
		} else if errors.Is(err, errutil.ErrCantDeleteOthersComment) {
			slog.DebugContext(ctx, "can't delete other's comment", slog.String("slug", slug), slog.String("commentId", commentId.String()), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusForbidden, "forbidden")
			return
		} else {
			ToInternalServerHTTPError(w, err)
			return
		}
	}
	ToSuccessHTTPResponse(w, nil)
}
