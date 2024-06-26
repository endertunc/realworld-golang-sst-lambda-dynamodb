package article

import (
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
)

type ArticlesApi struct {
}

func (aa ArticlesApi) GetArticles(w http.ResponseWriter, r *http.Request, params api.GetArticlesParams) {
	//TODO implement me
	panic("implement me")
}

func (aa ArticlesApi) CreateArticle(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (aa ArticlesApi) GetArticlesFeed(w http.ResponseWriter, r *http.Request, params api.GetArticlesFeedParams) {
	//TODO implement me
	panic("implement me")
}

func (aa ArticlesApi) DeleteArticle(w http.ResponseWriter, r *http.Request, slug string) {
	//TODO implement me
	panic("implement me")
}

func (aa ArticlesApi) GetArticle(w http.ResponseWriter, r *http.Request, slug string) {
	//TODO implement me
	panic("implement me")
}

func (aa ArticlesApi) UpdateArticle(w http.ResponseWriter, r *http.Request, slug string) {
	//TODO implement me
	panic("implement me")
}

func (aa ArticlesApi) DeleteArticleFavorite(w http.ResponseWriter, r *http.Request, slug string) {
	//TODO implement me
	panic("implement me")
}

func (aa ArticlesApi) CreateArticleFavorite(w http.ResponseWriter, r *http.Request, slug string) {
	//TODO implement me
	panic("implement me")
}
