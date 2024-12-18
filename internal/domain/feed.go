package domain

type ArticleAggregateView struct {
	Article     Article
	Author      User
	IsFollowing bool
	IsFavorited bool
}
