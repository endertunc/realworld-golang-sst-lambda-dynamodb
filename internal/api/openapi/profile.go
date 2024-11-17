package openapi

import (
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

type profileReq struct {
	Path string `path:"username"`
}

func buildProfile(reflector *openapi3.Reflector) {

	// GET /profiles/{username}
	type getProfileResp struct {
		profileReq
	}
	getProfileOp, _ := reflector.NewOperationContext(http.MethodGet, "/profiles/{username}")
	getProfileOp.AddReqStructure(new(getProfileResp))
	getProfileOp.AddRespStructure(new(dto.ProfileResponseBodyDTO), openapi.WithHTTPStatus(http.StatusOK))
	getProfileOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	getProfileOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	getProfileOp.AddSecurity(BearerAuthSecurityName)
	getProfileOp.AddSecurity(NoAuthSecurityName)
	_ = reflector.AddOperation(getProfileOp)

	// POST /profiles/{username}/follow
	type followProfileReq struct {
		profileReq
	}
	followProfileOp, _ := reflector.NewOperationContext(http.MethodPost, "/profiles/{username}/follow")
	followProfileOp.AddReqStructure(new(followProfileReq))
	followProfileOp.AddRespStructure(new(dto.ProfileResponseBodyDTO), openapi.WithHTTPStatus(http.StatusOK))
	followProfileOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	followProfileOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	followProfileOp.AddSecurity(BearerAuthSecurityName)
	_ = reflector.AddOperation(followProfileOp)

	// DELETE /profiles/{username}/follow
	type unfollowProfileReq struct {
		profileReq
	}
	unfollowProfileOp, _ := reflector.NewOperationContext(http.MethodDelete, "/profiles/{username}/follow")
	unfollowProfileOp.AddReqStructure(new(unfollowProfileReq))
	unfollowProfileOp.AddRespStructure(new(dto.ProfileResponseBodyDTO), openapi.WithHTTPStatus(http.StatusOK))
	unfollowProfileOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	unfollowProfileOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	unfollowProfileOp.AddSecurity(BearerAuthSecurityName)
	_ = reflector.AddOperation(unfollowProfileOp)
}
