package user

import (
	"context"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

type ArticleApi struct {
	ArticleService  ArticleServiceInterface
	UserService     UserServiceInterface
	FollowerService FollowerServiceInterface
}

type ArticleService struct {
	UserService       UserServiceInterface
	ArticleRepository ArticleRepositoryInterface
}

type ArticleServiceInterface interface {
	GetArticle(c context.Context, slug string) (domain.Article, error)
	CreateArticle(c context.Context, author uuid.UUID, title, description, body string, tagList []string) (domain.Article, error)
	//UpdateArticle(c context.Context, loggedInUserId uuid.UUID) (domain.Token, domain.User, error)
	AddComment(c context.Context, loggedInUserId uuid.UUID, articleSlug string, body string) (domain.Comment, error)
	GetArticleComments(c context.Context, slug string) ([]domain.Comment, error)
	DeleteComment(c context.Context, author uuid.UUID, slug string, commentId uuid.UUID) error
	DeleteArticle(c context.Context, author uuid.UUID, slug string) error
	FavoriteArticle(c context.Context, userId uuid.UUID, slug string) (domain.Article, error)
	UnfavoriteArticle(c context.Context, userId uuid.UUID, slug string) (domain.Article, error)
}

var _ ArticleServiceInterface = ArticleService{}

func (aa ArticleApi) GetArticle(c context.Context, slug string) (domain.Article, error) {
	article, err := aa.ArticleService.GetArticle(c, slug)
	if err != nil {
		return domain.Article{}, err
	}
	return article, nil
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

	_, user, err := aa.UserService.GetCurrentUser(c, loggedInUserId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// the current user is the author, and the user can't follow itself thus we simply pass isFollowing as false
	// the article has just been created thus we simply pass isFavorited ass false
	return dto.ToArticleResponseBodyDTO(article, user, false, false), nil
}

//func (aa ArticleApi) CreateComment(c context.Context, articleSlug string, author uuid.UUID, body string) (domain.Comment, domain.User, error) {
//
//	article, err := aa.ArticleService.CreateComment(c, articleSlug)
//	if err != nil {
//		return domain.Comment{}, domain.User{}, err
//	}
//	now := time.Now()
//
//	comment := domain.Comment{
//		Id:        uuid.New(),
//		ArticleId: article.Id,
//		AuthorId:  author,
//		Body:      body,
//		CreatedAt: now,
//		UpdatedAt: now,
//	}
//
//	as.ArticleRepository
//
//
//	return comment,
//
//}

func (aa ArticleApi) UnfavoriteArticle(c context.Context, loggedInUserId uuid.UUID, slug string) (dto.ArticleResponseBodyDTO, error) {
	article, err := aa.ArticleService.UnfavoriteArticle(c, loggedInUserId, slug)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	_, user, err := aa.UserService.GetCurrentUser(c, loggedInUserId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// ToDo @ender test if the parameters passed to isFollowing are correct
	isFollowing, err := aa.FollowerService.IsFollowing(c, loggedInUserId, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	return dto.ToArticleResponseBodyDTO(article, user, false, isFollowing), nil
}

func (aa ArticleApi) FavoriteArticle(c context.Context, loggedInUserId uuid.UUID, slug string) (dto.ArticleResponseBodyDTO, error) {
	article, err := aa.ArticleService.FavoriteArticle(c, loggedInUserId, slug)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}
	_, user, err := aa.UserService.GetCurrentUser(c, loggedInUserId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// ToDo @ender test if the parameters passed to isFollowing are correct
	isFollowing, err := aa.FollowerService.IsFollowing(c, loggedInUserId, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	return dto.ToArticleResponseBodyDTO(article, user, true, isFollowing), nil
}

func (aa ArticleApi) DeleteComment(c context.Context, loggedInUserId uuid.UUID, slug string, commentId uuid.UUID) error {
	err := aa.ArticleService.DeleteComment(c, loggedInUserId, slug, commentId)
	if err != nil {
		return err
	}
	return nil
}

func (aa ArticleApi) DeleteArticle(c context.Context, loggedInUserId uuid.UUID, slug string) error {
	err := aa.ArticleService.DeleteArticle(c, loggedInUserId, slug)
	if err != nil {
		return err
	}
	return nil
}

func (aa ArticleApi) GetArticleComments(c context.Context, loggedInUserId *uuid.UUID, slug string) (dto.MultiCommentsResponseBodyDTO, error) {
	comments, err := aa.ArticleService.GetArticleComments(c, slug)
	if err != nil {
		return dto.MultiCommentsResponseBodyDTO{}, err
	}

	// ToDo @ender we would like to extract unique author ids from comments
	// 	 check if we can simplify this later
	authorIdsMap := make(map[uuid.UUID]bool) // New empty set
	for _, comment := range comments {
		authorIdsMap[comment.AuthorId] = true
	}

	authorIdsList := make([]uuid.UUID, 0, len(authorIdsMap))
	for k := range authorIdsMap {
		authorIdsList = append(authorIdsList, k)
	}

	authors, err := aa.UserService.GetUserListByUserIDs(c, authorIdsList)

	if err != nil {
		return dto.MultiCommentsResponseBodyDTO{}, err
	}

	authorIdsToAuthorMap := make(map[uuid.UUID]domain.User, len(authors))
	for _, author := range authors {
		authorIdsToAuthorMap[author.Id] = author
	}
	// ToDo @ender we need to populate isFollowingMap from database
	return dto.ToMultiCommentsResponseBodyDTO(comments, authorIdsToAuthorMap, nil), nil
}

func (aa ArticleApi) AddComment(c context.Context, loggedInUserId uuid.UUID, articleSlug string, addCommentRequestDTO dto.AddCommentRequestBodyDTO) (dto.SingleCommentResponseBodyDTO, error) {

	comment, err := aa.ArticleService.AddComment(c, loggedInUserId, articleSlug, addCommentRequestDTO.Comment.Body)
	if err != nil {
		return dto.SingleCommentResponseBodyDTO{}, err
	}

	_, user, err := aa.UserService.GetCurrentUser(c, loggedInUserId)
	if err != nil {
		return dto.SingleCommentResponseBodyDTO{}, err
	}
	// the current user is the author, and the user can't follow itself,
	// thus we simply pass isFollowing as false
	return dto.ToSingleCommentResponseBodyDTO(comment, user, false), nil
}
