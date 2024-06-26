package comment

import (
	"net/http"
)

type CommentsApi struct {
}

func (CommentsApi) GetArticleComments(w http.ResponseWriter, r *http.Request, slug string) {
	//TODO implement me
	panic("implement me")
}

func (CommentsApi) CreateArticleComment(w http.ResponseWriter, r *http.Request, slug string) {
	//TODO implement me
	panic("implement me")
}

func (CommentsApi) DeleteArticleComment(w http.ResponseWriter, r *http.Request, slug string, id int) {
	//TODO implement me
	panic("implement me")
}
