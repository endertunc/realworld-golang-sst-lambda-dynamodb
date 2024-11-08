package user

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

type ArticleApi struct {
	ArticleService ArticleServiceInterface
	UserService    UserServiceInterface
	ProfileService ProfileServiceInterface
}

func NewArticleApi(articleService ArticleServiceInterface, userService UserServiceInterface, profileService ProfileServiceInterface) ArticleApi {
	return ArticleApi{
		ArticleService: articleService,
		UserService:    userService,
		ProfileService: profileService,
	}
}

type ArticleService struct {
	UserService       UserServiceInterface
	ArticleRepository ArticleRepositoryInterface
}

type ArticleServiceInterface interface {
	GetArticle(c context.Context, slug string) (domain.Article, error)
	CreateArticle(c context.Context, author uuid.UUID, title, description, body string, tagList []string) (domain.Article, error)
	AddComment(c context.Context, loggedInUserId uuid.UUID, articleSlug string, body string) (domain.Comment, error)
	GetArticleComments(c context.Context, slug string) ([]domain.Comment, error)
	DeleteComment(c context.Context, author uuid.UUID, slug string, commentId uuid.UUID) error
	DeleteArticle(c context.Context, author uuid.UUID, slug string) error
	FavoriteArticle(c context.Context, userId uuid.UUID, slug string) (domain.Article, error)
	UnfavoriteArticle(c context.Context, userId uuid.UUID, slug string) (domain.Article, error)
	IsFavorited(c context.Context, articleId, userId uuid.UUID) (bool, error)
	FindArticlesByIds(c context.Context, articleIds []uuid.UUID) ([]domain.Article, error)
	IsFavoritedBulk(c context.Context, userId uuid.UUID, articleIds []uuid.UUID) (map[uuid.UUID]bool, error)
	//UpdateArticle(c context.Context, loggedInUserId uuid.UUID) (domain.Token, domain.User, error)
}

var _ ArticleServiceInterface = ArticleService{}

func (aa ArticleApi) GetArticle(ctx context.Context, loggedInUserId *uuid.UUID, slug string) (dto.ArticleResponseBodyDTO, error) {
	// ToDo @ender [maybe not] --- we could get article and author in BatchGetItem
	article, err := aa.ArticleService.GetArticle(ctx, slug)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}
	author, err := aa.UserService.GetUserByUserId(ctx, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}
	if loggedInUserId == nil {
		return dto.ToArticleResponseBodyDTO(article, author, false, false), nil
	} else {
		loggedInUser, err := aa.UserService.GetCurrentUser(ctx, *loggedInUserId)
		if err != nil {
			return dto.ArticleResponseBodyDTO{}, err
		}

		// ToDo @ender we make multiple request. We could optimize this by using BatchGetItem - isFollowing and isFavorited
		isFollowing, err := aa.ProfileService.IsFollowing(ctx, loggedInUser.Id, article.AuthorId)
		if err != nil {
			return dto.ArticleResponseBodyDTO{}, err
		}

		isFavorited, err := aa.ArticleService.IsFavorited(ctx, article.Id, loggedInUser.Id)

		if err != nil {
			return dto.ArticleResponseBodyDTO{}, err
		}

		return dto.ToArticleResponseBodyDTO(article, author, isFavorited, isFollowing), nil
	}

}

func (aa ArticleApi) CreateArticle(c context.Context, loggedInUserId uuid.UUID, createArticleRequestBodyDTO dto.CreateArticleRequestBodyDTO) (dto.ArticleResponseBodyDTO, error) {
	articleDTO := createArticleRequestBodyDTO.Article
	// ToDo @ender do we have any business validation we should apply in service level for an article?
	// ToDo @ender [GENERAL] - in this project we don't seem to have much complex data types to pass to services
	//  thus I skipped creating a struct that "service accepts" and simply passed the params needed to create and article
	//  Once this list of parameters that needs to be passed to service gets crowded,
	//  one could introduce intermediate "CreateArticleRequest" that ArticleService accepts
	article, err := aa.ArticleService.CreateArticle(
		c,
		loggedInUserId,
		articleDTO.Title,
		articleDTO.Description,
		articleDTO.Body,
		articleDTO.TagList)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	user, err := aa.UserService.GetCurrentUser(c, loggedInUserId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// the current user is the author, and the user can't follow itself thus we simply pass isFollowing as false
	// the article has just been created thus we simply pass isFavorited as false
	return dto.ToArticleResponseBodyDTO(article, user, false, false), nil
}

func (aa ArticleApi) UnfavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) (dto.ArticleResponseBodyDTO, error) {
	article, err := aa.ArticleService.UnfavoriteArticle(ctx, loggedInUserId, slug)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	author, err := aa.UserService.GetUserByUserId(ctx, loggedInUserId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// ToDo @ender test if the parameters passed to isFollowing are correct
	isFollowing, err := aa.ProfileService.IsFollowing(ctx, loggedInUserId, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	return dto.ToArticleResponseBodyDTO(article, author, false, isFollowing), nil
}

func (aa ArticleApi) FavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) (dto.ArticleResponseBodyDTO, error) {
	article, err := aa.ArticleService.FavoriteArticle(ctx, loggedInUserId, slug)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	author, err := aa.UserService.GetUserByUserId(ctx, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// ToDo @ender test if the parameters passed to isFollowing are correct
	isFollowing, err := aa.ProfileService.IsFollowing(ctx, loggedInUserId, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	return dto.ToArticleResponseBodyDTO(article, author, true, isFollowing), nil
}

func (aa ArticleApi) DeleteComment(ctx context.Context, loggedInUserId uuid.UUID, slug string, commentId uuid.UUID) error {
	err := aa.ArticleService.DeleteComment(ctx, loggedInUserId, slug, commentId)
	if err != nil {
		return err
	}
	return nil
}

func (aa ArticleApi) DeleteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) error {
	err := aa.ArticleService.DeleteArticle(ctx, loggedInUserId, slug)
	if err != nil {
		return err
	}
	return nil
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
func (aa ArticleApi) GetArticleComments(ctx context.Context, loggedInUserId *uuid.UUID, slug string) (dto.MultiCommentsResponseBodyDTO, error) {
	comments, err := aa.ArticleService.GetArticleComments(ctx, slug)
	if err != nil {
		return dto.MultiCommentsResponseBodyDTO{}, err
	}
	slog.DebugContext(ctx, "comments after ArticleService.GetArticleComments", slog.Any("comments", comments))
	if len(comments) == 0 {
		return dto.MultiCommentsResponseBodyDTO{Comment: []dto.CommentResponseDTO{}}, nil
	} else {
		// ToDo @ender we would like to extract unique author ids from comments
		// 	 check if we can simplify this later
		authorIdsMap := make(map[uuid.UUID]bool) // New empty set
		for _, comment := range comments {
			authorIdsMap[comment.AuthorId] = true
		}

		uniqueAuthorIdsList := make([]uuid.UUID, 0, len(authorIdsMap))
		for k := range authorIdsMap {
			uniqueAuthorIdsList = append(uniqueAuthorIdsList, k)
		}

		authors, err := aa.UserService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)

		if err != nil {
			return dto.MultiCommentsResponseBodyDTO{}, err
		}

		authorIdsToAuthorMap := make(map[uuid.UUID]domain.User, len(authors))
		for _, author := range authors {
			authorIdsToAuthorMap[author.Id] = author
		}

		if loggedInUserId == nil {
			return dto.ToMultiCommentsResponseBodyDTO(comments, authorIdsToAuthorMap, map[uuid.UUID]bool{}), nil
		} else {
			followedAuthorsMap, err := aa.ProfileService.IsFollowingBulk(ctx, *loggedInUserId, uniqueAuthorIdsList)
			if err != nil {
				return dto.MultiCommentsResponseBodyDTO{}, err
			}
			return dto.ToMultiCommentsResponseBodyDTO(comments, authorIdsToAuthorMap, followedAuthorsMap), nil
		}

	}
}

func (aa ArticleApi) AddComment(ctx context.Context, loggedInUserId uuid.UUID, articleSlug string, addCommentRequestDTO dto.AddCommentRequestBodyDTO) (dto.SingleCommentResponseBodyDTO, error) {
	comment, err := aa.ArticleService.AddComment(ctx, loggedInUserId, articleSlug, addCommentRequestDTO.Comment.Body)
	if err != nil {
		return dto.SingleCommentResponseBodyDTO{}, err
	}

	user, err := aa.UserService.GetCurrentUser(ctx, loggedInUserId)
	if err != nil {
		return dto.SingleCommentResponseBodyDTO{}, err
	}
	// the current user is the author, and the user can't follow itself,
	// thus we simply pass isFollowing as false
	return dto.ToSingleCommentResponseBodyDTO(comment, user, false), nil
}

func (as ArticleService) IsFavorited(c context.Context, articleId, userId uuid.UUID) (bool, error) {
	return as.ArticleRepository.IsFavorited(c, articleId, userId)
}

func (as ArticleService) FindArticlesByIds(c context.Context, articleIds []uuid.UUID) ([]domain.Article, error) {
	return as.ArticleRepository.FindArticlesByIds(c, articleIds)
}

func (as ArticleService) IsFavoritedBulk(c context.Context, userId uuid.UUID, articleIds []uuid.UUID) (map[uuid.UUID]bool, error) {
	return as.ArticleRepository.IsFavoritedBulk(c, userId, articleIds)
}