package domain

// ToDo @ender better name for this struct
type FeedItem struct {
	Article     Article
	Author      User
	IsFollowing bool
	IsFavorited bool
}
