package article

type ArticleService struct {
}

type ArticleServiceInterface interface {
	CreateArticle() bool
}

func (s ArticleService) CreateArticle() bool {
	return true
}
