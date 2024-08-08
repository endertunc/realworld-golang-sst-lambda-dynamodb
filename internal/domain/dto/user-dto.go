package dto

import "realworld-aws-lambda-dynamodb-golang/internal/domain"

type LoginRequestBodyDTO struct {
	User LoginRequestUserDto `json:"user"`
}

type LoginRequestUserDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponseBodyDTO struct {
	User UserResponseUserDto `json:"user"`
}
type UserResponseUserDto struct {
	Email    string  `json:"email"`
	Token    string  `json:"password"`
	Username string  `json:"username"`
	Bio      *string `json:"bio"`
	Image    *string `json:"image"`
}

func ToUserResponseBodyDTO(user domain.User, token domain.Token) UserResponseBodyDTO {
	return UserResponseBodyDTO{
		User: UserResponseUserDto{
			Email:    user.Email,
			Token:    string(token),
			Username: user.Username,
			Bio:      user.Bio,
			Image:    user.Image,
		},
	}
}

type NewUserRequestBodyDTO struct {
	User NewUserRequestUserDto `json:"user"`
}

type NewUserRequestUserDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}
