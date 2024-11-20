package domain

// ToDo @ender better name for this struct: an idea ArticleAggregateItem
type FeedItem struct {
	Article     Article
	Author      User
	IsFollowing bool
	IsFavorited bool
}
