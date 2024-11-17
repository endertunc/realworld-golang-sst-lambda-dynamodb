package openapi

import (
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

type commentReq struct {
	Path string `path:"slug"`
}

func buildComment(reflector *openapi3.Reflector) {

	// GET /articles/{slug}/comments
	type getCommentsReq struct {
		commentReq
	}
	getCommentsOp, _ := reflector.NewOperationContext(http.MethodGet, "/articles/{slug}/comments")
	getCommentsOp.AddReqStructure(new(getCommentsReq))
	getCommentsOp.AddRespStructure(new(dto.MultiCommentsResponseBodyDTO))
	getCommentsOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusNotFound))
	getCommentsOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	getCommentsOp.AddSecurity(BearerAuthSecurityName)
	getCommentsOp.AddSecurity(NoAuthSecurityName)
	_ = reflector.AddOperation(getCommentsOp)

	// POST /articles/{slug}/comments
	type addCommentReq struct {
		commentReq
	}
	addCommentOp, _ := reflector.NewOperationContext(http.MethodPost, "/articles/{slug}/comments")
	addCommentOp.AddReqStructure(new(addCommentReq))
	addCommentOp.AddReqStructure(new(dto.AddCommentRequestBodyDTO))
	addCommentOp.AddRespStructure(new(dto.SingleCommentResponseBodyDTO))
	addCommentOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	addCommentOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusNotFound))
	addCommentOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	addCommentOp.AddSecurity(BearerAuthSecurityName)
	_ = reflector.AddOperation(addCommentOp)

	// DELETE /articles/{slug}/comments/{id}
	type deleteCommentReq struct {
		commentReq
	}
	deleteCommentOp, _ := reflector.NewOperationContext(http.MethodDelete, "/articles/{slug}/comments/{id}")
	deleteCommentOp.AddReqStructure(new(deleteCommentReq))
	deleteCommentOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	deleteCommentOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusNotFound))
	deleteCommentOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	deleteCommentOp.AddSecurity(BearerAuthSecurityName)
	_ = reflector.AddOperation(deleteCommentOp)
}
