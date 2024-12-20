package errutil

import (
	"errors"
)

type SimpleError struct {
	Message string `json:"message"`
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// note on where to define errors (and I am not still sure which one I like better):
// - define errors in where they are returned: errors are closely related to the functions that return them.
// - define errors in a single place: easy to find and manage all errors in one place.
var (
	ErrJsonDecode            = errors.New("json decode failed")
	ErrJsonEncode            = errors.New("json encode failed")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrArticleNotFound       = errors.New("article not found")
	// ErrHashPassword will be mapped to InternalServerError anyway, so I might as well remove this.
	ErrHashPassword            = errors.New("hash password failed")
	ErrTokenGenerate           = errors.New("generate token failed")
	ErrDynamoQuery             = errors.New("dynamodb query failed")
	ErrDynamoMapping           = errors.New("dynamodb mapping failed")
	ErrDynamoMarshalling       = errors.New("dynamodb marshalling failed")
	ErrOpensearchMarshalling   = errors.New("opensearch marshalling failed")
	ErrOpensearchQuery         = errors.New("opensearch query failed")
	ErrDynamoTokenDecoding     = errors.New("dynamodb token decoding failed")
	ErrDynamoTokenEncoding     = errors.New("dynamodb token encoding failed")
	ErrCantFollowYourself      = errors.New("cannot follow yourself")
	ErrCantDeleteOthersComment = errors.New("cannot delete other's comment")
	ErrCantDeleteOthersArticle = errors.New("cannot delete other's article")
	ErrCantUpdateOthersArticle = errors.New("cannot update other's article")
	ErrCommentNotFound         = errors.New("comment not found")
	ErrAlreadyFavorited        = errors.New("already favorited")
	ErrAlreadyUnfavorited      = errors.New("already unfavorited")
	ErrSlugAlreadyExists       = errors.New("slug already exists")
)
