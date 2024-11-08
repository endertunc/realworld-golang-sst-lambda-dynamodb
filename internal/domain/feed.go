package domain

type FeedItem struct {
	Article     Article
	Author      User
	IsFollowing bool
	IsFavorited bool
}
