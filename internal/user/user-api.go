package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

type UserApi struct {
	UserService UserServiceInterface
}

//func (ua UserApi) Login(ctx context.Context, request api.LoginRequestObject) (api.LoginResponseObject, error) {
//	token, user, err := ua.UserService.LoginUser(ctx, request.Body.User.Email, request.Body.User.Password)
//	if err != nil {
//		// ToDo log error...
//		//invalidPasswordError := &domain.InvalidPasswordError{}
//		//if errors.As(err, &invalidPasswordError) {
//		//	return api.Login401Response{}, nil
//		//}
//		//
//		//userNotFoundError := &domain.UserNotFoundError{}
//		//if errors.As(err, &userNotFoundError) {
//		//	return api.Login401Response{}, nil
//		//}
//
//		return api.Login422JSONResponse{
//			GenericErrorJSONResponse: api.ToGenericErrorResponse(err),
//		}, nil
//	}
//
//	// ToDo find a way to convert *T to nullable.Nullable[T]
//	bio := nullable.NewNullNullable[string]()
//	if user.Bio != nil {
//		bio = nullable.NewNullableWithValue[string](*user.Bio)
//	}
//
//	// ToDo find a way to convert *T to nullable.Nullable[T]
//	image := nullable.NewNullNullable[string]()
//	if user.Image != nil {
//		image = nullable.NewNullableWithValue[string](*user.Image)
//	}
//
//	userDTO := api.User{
//		Bio:      bio,
//		Email:    user.Email,
//		Image:    image,
//		Token:    string(*token),
//		Username: user.Username,
//	}
//
//	return api.Login200JSONResponse{
//		UserResponseJSONResponse: api.UserResponseJSONResponse{
//			User: userDTO,
//		},
//	}, nil
//
//}

func (ua UserApi) GetProfileByUsername(w http.ResponseWriter, r *http.Request, username string) {
	//TODO implement me
	panic("implement me")
}

func (ua UserApi) GetTags(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (ua UserApi) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (ua UserApi) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (ua UserApi) CreateUser(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (ua UserApi) Login(w http.ResponseWriter, r *http.Request) {
	loginJSONRequestBody := api.LoginJSONRequestBody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&loginJSONRequestBody)
	if err != nil {
		cause := fmt.Errorf("user-api.login - error decoding request body: %w", err)
		errutil.WriteToResponse(errutil.BadRequestError("error decoding request body", cause), w)
		return
	}
	// ToDo validate login request body!!!
	loginUser := loginJSONRequestBody.User
	token, user, err := ua.UserService.LoginUser(r.Context(), loginUser.Email, loginUser.Password)
	if err != nil {
		errutil.WriteToResponse(err, w)
		return
	}
	userDTO := api.User{
		Email:    user.Email,
		Username: user.Username,
		Token:    string(*token),
		Bio:      user.Bio,
		Image:    user.Image,
	}
	err = json.NewEncoder(w).Encode(api.UserResponse{User: userDTO})
	if err != nil {
		cause := fmt.Errorf("user-api.login - error encoding response body: %w", err)
		errutil.WriteToResponse(errutil.InternalError(cause), w)
		return
	}
}
