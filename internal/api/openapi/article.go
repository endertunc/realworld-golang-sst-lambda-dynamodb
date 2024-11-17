package openapi

import (
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

type queryParameterLimit struct {
	Limit int `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

type queryParameterOffset struct {
	Offset string `query:"offset"`
}

type articleReq struct {
	Path string `path:"slug"`
}

func buildArticle(reflector *openapi3.Reflector) {
	// GET /articles/feed
	type getFeedReq struct {
		queryParameterLimit
		queryParameterOffset
	}

	getFeedOp, _ := reflector.NewOperationContext(http.MethodGet, "/articles/feed")
	getFeedOp.AddReqStructure(new(getFeedReq))
	getFeedOp.AddRespStructure(new(dto.MultipleArticlesResponseBodyDTO), openapi.WithHTTPStatus(http.StatusOK))
	getFeedOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	getFeedOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	getFeedOp.AddSecurity(BearerAuthSecurityName)
	_ = reflector.AddOperation(getFeedOp)

	// GET /articles
	type getArticlesReq struct {
		Author    string `query:"author"`
		Favorited string `query:"favorited"`
		Tag       string `query:"tag"`
		queryParameterLimit
		queryParameterOffset
	}

	getArticlesOp, _ := reflector.NewOperationContext(http.MethodGet, "/articles")
	getArticlesOp.AddReqStructure(new(getArticlesReq))
	getArticlesOp.AddRespStructure(new(dto.MultipleArticlesResponseBodyDTO), openapi.WithHTTPStatus(http.StatusOK))
	getArticlesOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	getArticlesOp.AddSecurity(BearerAuthSecurityName)
	getArticlesOp.AddSecurity(NoAuthSecurityName)

	_ = reflector.AddOperation(getArticlesOp)

	// POST /articles
	createArticleOp, _ := reflector.NewOperationContext(http.MethodPost, "/articles")
	createArticleOp.AddReqStructure(new(dto.CreateArticleRequestBodyDTO))
	createArticleOp.AddRespStructure(new(dto.ArticleResponseBodyDTO), openapi.WithHTTPStatus(http.StatusCreated))
	createArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	createArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	createArticleOp.AddSecurity(BearerAuthSecurityName)
	_ = reflector.AddOperation(createArticleOp)

	// GET /articles/{slug}
	type getArticleReq struct {
		articleReq
	}
	getArticleOp, _ := reflector.NewOperationContext(http.MethodGet, "/articles/{slug}")
	getArticleOp.AddReqStructure(new(getArticleReq))
	getArticleOp.AddRespStructure(new(dto.ArticleResponseBodyDTO), openapi.WithHTTPStatus(http.StatusOK))
	getArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusNotFound))
	getArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	getArticleOp.AddSecurity(BearerAuthSecurityName)
	getArticleOp.AddSecurity(NoAuthSecurityName)
	_ = reflector.AddOperation(getArticleOp)

	// PUT /articles/{slug} ToDo

	// DELETE /articles/{slug}
	type deleteArticleReq struct {
		articleReq
	}
	deleteArticleOp, _ := reflector.NewOperationContext(http.MethodDelete, "/articles/{slug}")
	deleteArticleOp.AddReqStructure(new(deleteArticleReq))
	deleteArticleOp.AddRespStructure(nil, openapi.WithHTTPStatus(http.StatusNoContent))
	deleteArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusNotFound))
	deleteArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	deleteArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	deleteArticleOp.AddSecurity(BearerAuthSecurityName)
	_ = reflector.AddOperation(deleteArticleOp)

	// POST /articles/{slug}/favorite
	type favoriteArticleReq struct {
		articleReq
	}
	favoriteArticleOp, _ := reflector.NewOperationContext(http.MethodPost, "/articles/{slug}/favorite")
	favoriteArticleOp.AddReqStructure(new(favoriteArticleReq))
	favoriteArticleOp.AddRespStructure(new(dto.ArticleResponseBodyDTO))
	favoriteArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	favoriteArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusNotFound))
	favoriteArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	favoriteArticleOp.AddSecurity(BearerAuthSecurityName)
	_ = reflector.AddOperation(favoriteArticleOp)

	// DELETE /articles/{slug}/favorite
	type unfavoriteArticleReq struct {
		articleReq
	}
	unfavoriteArticleOp, _ := reflector.NewOperationContext(http.MethodDelete, "/articles/{slug}/favorite")
	unfavoriteArticleOp.AddReqStructure(new(unfavoriteArticleReq))
	unfavoriteArticleOp.AddRespStructure(new(dto.ArticleResponseBodyDTO))
	unfavoriteArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	unfavoriteArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusNotFound))
	unfavoriteArticleOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	unfavoriteArticleOp.AddSecurity(BearerAuthSecurityName)
	_ = reflector.AddOperation(unfavoriteArticleOp)
}
