package openapi

import (
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

func buildUser(reflector *openapi3.Reflector) {

	// POST /users/login
	loginOp, _ := reflector.NewOperationContext(http.MethodPost, "/users/login")
	loginOp.AddReqStructure(new(dto.NewUserRequestBodyDTO))
	loginOp.AddRespStructure(new(dto.UserResponseBodyDTO), openapi.WithHTTPStatus(http.StatusOK))
	loginOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = reflector.AddOperation(loginOp)

	// POST /users
	createUserOp, _ := reflector.NewOperationContext(http.MethodPost, "/users")
	createUserOp.AddReqStructure(new(dto.NewUserRequestBodyDTO))
	createUserOp.AddRespStructure(new(dto.UserResponseBodyDTO), openapi.WithHTTPStatus(http.StatusCreated))
	createUserOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusConflict))
	createUserOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnprocessableEntity))
	_ = reflector.AddOperation(createUserOp)

	// GET /user
	getCurrentUserOp, _ := reflector.NewOperationContext(http.MethodGet, "/user")
	getCurrentUserOp.AddRespStructure(new(dto.UserResponseBodyDTO), openapi.WithHTTPStatus(http.StatusOK))
	getCurrentUserOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusUnauthorized))
	getCurrentUserOp.AddRespStructure(new(errutil.SimpleError), openapi.WithHTTPStatus(http.StatusInternalServerError))
	getCurrentUserOp.AddSecurity(BearerAuthSecurityName)
	_ = reflector.AddOperation(getCurrentUserOp)

	// PUT /user TODO @ender

}
