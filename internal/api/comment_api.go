package api

import (
	"context"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
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
func (aa CommentApi) GetArticleComments(ctx context.Context, loggedInUserId *uuid.UUID, slug string) (dto.MultiCommentsResponseBodyDTO, error) {
	comments, err := aa.commentService.GetArticleComments(ctx, slug)
	if err != nil {
		return dto.MultiCommentsResponseBodyDTO{}, err
	}
	if len(comments) == 0 {
		return dto.MultiCommentsResponseBodyDTO{Comment: []dto.CommentResponseDTO{}}, nil
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
			return dto.MultiCommentsResponseBodyDTO{}, err
		}

		authorIdsToAuthorMap := make(map[uuid.UUID]domain.User, len(authors))
		for _, author := range authors {
			authorIdsToAuthorMap[author.Id] = author
		}

		if loggedInUserId == nil {
			return dto.ToMultiCommentsResponseBodyDTO(comments, authorIdsToAuthorMap, mapset.NewSetWithSize[uuid.UUID](0)), nil
		} else {
			followedAuthorsSet, err := aa.profileService.IsFollowingBulk(ctx, *loggedInUserId, uniqueAuthorIdsList)
			if err != nil {
				return dto.MultiCommentsResponseBodyDTO{}, err
			}
			return dto.ToMultiCommentsResponseBodyDTO(comments, authorIdsToAuthorMap, followedAuthorsSet), nil
		}

	}
}

func (aa CommentApi) AddComment(ctx context.Context, loggedInUserId uuid.UUID, articleSlug string, addCommentRequestDTO dto.AddCommentRequestBodyDTO) (dto.SingleCommentResponseBodyDTO, error) {
	comment, err := aa.commentService.AddComment(ctx, loggedInUserId, articleSlug, addCommentRequestDTO.Comment.Body)
	if err != nil {
		return dto.SingleCommentResponseBodyDTO{}, err
	}

	user, err := aa.userService.GetUserByUserId(ctx, loggedInUserId)
	if err != nil {
		return dto.SingleCommentResponseBodyDTO{}, err
	}
	// the current user is the author, and the user can't follow itself,
	// thus we simply pass isFollowing as false
	return dto.ToSingleCommentResponseBodyDTO(comment, user, false), nil
}

func (aa CommentApi) DeleteComment(ctx context.Context, loggedInUserId uuid.UUID, slug string, commentId uuid.UUID) error {
	err := aa.commentService.DeleteComment(ctx, loggedInUserId, slug, commentId)
	if err != nil {
		return err
	}
	return nil
}
